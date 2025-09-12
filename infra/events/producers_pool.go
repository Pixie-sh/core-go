package events

import (
	"context"
	"slices"

	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/infra/message_factory"
	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	"github.com/pixie-sh/core-go/pkg/types"
)

const EventTypesWildcard = "*" //always triggered
const EventTypesFallback = "?" //triggered when there's no other rule

type ProducerPoolConfiguration struct {
	ProducerPoolID                    string              `json:"producer_pool_id"`
	SupportedPayloadTypesByProducerID map[string][]string `json:"supported_payload_types"`
	SupportedPacksByProducerID        map[string][]string `json:"supported_packs"`
}

type SupportedEvents struct {
	SupportedPayloadTypesByProducerID []string `json:"supported_payload_types"`
	SupportedPacksByProducerID        []string `json:"supported_packs"`
}

type ProducersPool struct {
	config          ProducerPoolConfiguration
	producersList   []Producer
	producersMapped map[string][]Producer //producersMapped is not thread safe; meant to be changed on constructor phase
}

func NewProducersPool(ctx context.Context, config ProducerPoolConfiguration, producers ...Producer) (ProducersPool, error) {
	pp := ProducersPool{
		config:          config,
		producersList:   producers,
		producersMapped: make(map[string][]Producer),
	}

	err := pp.digestConfiguration(ctx)
	pixiecontext.GetCtxLogger(ctx).
		With("producer_pool_id", pp.ID()).
		With("producer_pool_config", pp.config).
		Log("producers pool digested")
	return pp, err
}

func (p *ProducersPool) digestConfiguration(ctx context.Context) error {
	for _, producer := range p.producersList {
		supportedPayloadTypes := p.config.SupportedPayloadTypesByProducerID[producer.ID()]
		supportedPacks := p.config.SupportedPacksByProducerID[producer.ID()]
		if len(supportedPayloadTypes) == 0 && len(supportedPacks) == 0 {
			pixiecontext.GetCtxLogger(ctx).With("producer_id", producer.ID()).Warn("producer %s doesn't have supported payload types or Packs", producer.ID())
			continue
		}

		if validateWildcardConfiguration(ctx, supportedPayloadTypes) || validateWildcardConfiguration(ctx, supportedPacks) {
			p.producersMapped[EventTypesWildcard] = append(p.producersMapped[EventTypesWildcard], producer)
			continue //if producer allows all (EventTypesWildcard), don't need the specific; otherwise will be repeated
		}

		for _, payloadType := range supportedPayloadTypes {

			if p.checkExistingEntry(ctx, producer, payloadType) {
				pixiecontext.GetCtxLogger(ctx).
					With("producer_id", producer.ID()).
					With("payload_type", payloadType).
					Warn("payload type already exists for this producer, will ignore")
				continue
			}

			p.producersMapped[payloadType] = append(p.producersMapped[payloadType], producer)
		}

		// Not sure if this could come from a different factory...
		packs := message_factory.Singleton.GetRegisteredPacks()

		for _, supportedPack := range supportedPacks {
			pack, ok := packs[supportedPack]
			if !ok {
				return errors.New("Pack %s does not exist.", supportedPack)
			}

			p.addPackEntries(ctx, producer, pack)
		}

	}

	pixiecontext.GetCtxLogger(ctx).Debug("Registered Producers: %+v", p.producersMapped)
	return nil
}

func validateWildcardConfiguration(_ context.Context, supportedEvents []string) bool {

	return slices.ContainsFunc(supportedEvents, func(s string) bool {
		return s == EventTypesWildcard
	})
}

func (p *ProducersPool) checkExistingEntry(_ context.Context, producer Producer, entry string) bool {
	return slices.ContainsFunc(p.producersMapped[entry], func(availableProducer Producer) bool {
		return availableProducer.ID() == producer.ID()
	})
}

func (p *ProducersPool) addPackEntries(ctx context.Context, producer Producer, pack message_factory.Pack) {

	for _, entryName := range pack.EntryNames() { // Damn, this is a lot of for loops

		if p.checkExistingEntry(ctx, producer, entryName) {
			pixiecontext.GetCtxLogger(ctx).
				With("producer_id", producer.ID()).
				With("payload_type", entryName).
				With("pack", pack.Name).
				Warn("payload type already exists for this producer, will ignore")
			continue
		}

		p.producersMapped[entryName] = append(p.producersMapped[entryName], producer)

	}

}

func (p *ProducersPool) ID() string {
	return p.config.ProducerPoolID
}

func (p *ProducersPool) ProduceBatch(ctx context.Context, wrappers ...UntypedEventWrapper) error {
	log := pixiecontext.GetCtxLogger(ctx)
	if len(wrappers) == 0 {
		log.Warn("provided event wrappers are empty")
		return errors.New("provided event wrappers are empty")
	}

	groups := make(map[string][]UntypedEventWrapper)
	for _, wrapper := range wrappers {
		if types.Nil(wrapper) {
			continue
		}

		pt := wrapper.PayloadType
		groups[pt] = append(groups[pt], wrapper)
	}

	var err []error
	for ptype, w := range groups {
		err = append(err, p.produceWithPayloadType(ctx, log, ptype, w...))
	}

	return errors.Join(err...)
}

func (p *ProducersPool) produceWithPayloadType(ctx context.Context, log logger.Interface, payloadType string, wrappers ...UntypedEventWrapper) error {
	producers, ok := p.producersMapped[payloadType]
	if !ok {
		log.Warn("no producers found for payload type: %s", payloadType)
	}

	producers = append(producers, p.producersMapped[EventTypesWildcard]...)
	if len(producers) == 0 {
		log.Warn("no wildcard producers found for payload type: %s", payloadType)
		return errors.New("no producers found for payload type '%s' nor for '%s'", payloadType, EventTypesWildcard)
	}

	log.Debug("Producers that contain payload type %s are: %+v", payloadType, producers)

	var errorsList []error
	var alreadyProducedBy []string
	for _, producer := range producers {
		log.Debug("using producer %s", producer.ID())
		if slices.Contains(alreadyProducedBy, producer.ID()) {
			continue
		}

		log.Debug("producing batch producer %s", producer.ID())
		err := producer.ProduceBatch(ctx, wrappers...)
		if err != nil {
			log.With("error", err).Error("failed to produce event: %s", err.Error())
			errorsList = append(errorsList, err)
			continue
		}

		alreadyProducedBy = append(alreadyProducedBy, producer.ID())
	}

	return errors.Join(errorsList...)
}

func (p *ProducersPool) Produce(ctx context.Context, wrapper UntypedEventWrapper) error {
	log := pixiecontext.GetCtxLogger(ctx)
	if types.IsEmpty(wrapper) {
		log.Warn("provided event wrapper is nil")
		return errors.New("provided event wrapper is nil")
	}

	producers, ok := p.producersMapped[wrapper.PayloadType]
	if !ok {
		log.Warn("no producers found for payload type: %s", wrapper.PayloadType)
	}

	log.Debug("Producers that contain payload type %s are: %+v", wrapper.PayloadType, producers)

	producers = append(producers, p.producersMapped[EventTypesWildcard]...)
	if len(producers) == 0 {
		log.Warn("no wildcard producers found for payload type: %s", wrapper.PayloadType)
		return errors.New("no producers found for payload type '%s' nor for '%s'", wrapper.PayloadType, EventTypesWildcard)
	}

	var errorsList []error
	var alreadyProducedBy []string
	for _, producer := range producers {
		if slices.Contains(alreadyProducedBy, producer.ID()) {
			continue
		}

		err := producer.Produce(ctx, wrapper)
		if err != nil {
			log.Error("failed to produce event: %s", err.Error())
			errorsList = append(errorsList, err)
			continue
		}

		alreadyProducedBy = append(alreadyProducedBy, producer.ID())
	}

	return errors.Join(errorsList...)
}
