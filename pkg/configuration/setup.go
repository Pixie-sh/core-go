package configuration

import (
	"flag"
	"os"
	"reflect"
	"strings"

	jsonpatch "github.com/evanphx/json-patch/v5"
	gojson "github.com/goccy/go-json"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"
	"github.com/wI2L/jsondiff"

	internalEnv "github.com/pixie-sh/core-go/pkg/env"
)

// Setup set up configuration from file.
// ignoreArgs=true allows the process to proceed even if no args were passed to the executable,
// using env variables
func Setup(cfg interface{}, ignoreArgs bool, withValidations ...bool) {
	var log logger.Interface = logger.Logger
	var configurationFile string
	var envFile string
	if !ignoreArgs {
		flagConfigurationFile := flag.Lookup("config")
		flagEnvFile := flag.Lookup("env")

		if flagConfigurationFile == nil {
			log.Error("unable to find flag -config")
			flag.PrintDefaults()
			os.Exit(1)
		}

		configurationFile = flagConfigurationFile.Value.String()

		// provided env file path is empty
		// let's try use default env variable for the filepath
		envFile = os.Getenv(internalEnv.EnvFile)
		if flagEnvFile != nil {
			envFile = flagEnvFile.Value.String()
		}

		err := SetEnvFromFile(envFile)
		if err != nil {
			log.Error("unable to load env from file (%s) error (%s) \n", envFile, err.Error())
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	var configFiles []string
	if len(configurationFile) == 0 {
		configsEnv := os.Getenv(internalEnv.EnvConfigs)
		if len(configsEnv) == 0 { //no config provided, use default
			log.Debug("using default config bootstrap.json")
			configFiles = append(configFiles, "bootstrap.json")
		} else {
			configFiles = append(configFiles, strings.Split(configsEnv, ",")...)
		}
	} else {
		configFiles = append(configFiles, configurationFile)
	}

	loaded := false
	var jsonBlob []byte
	var err error
	if len(configFiles) == 1 {
		jsonBlob, err = StructFromFileWithEnvReplace(configFiles[0], cfg, log)
		if err != nil {
			log.Error("unable to load config from file (%s) error (%s) \n", configFiles[0], err.Error())
			os.Exit(1)
		}

		loaded = true
	}

	if len(configFiles) > 1 && !loaded {
		// Multi-file mode: merge raw JSON first, then resolve envs/refs once on
		// the merged document. Per-file resolution would substitute the base's
		// ${#ref.X} placeholders against the base's own #ref table before
		// overlays could contribute alternate #ref.X blocks via merge.
		rawConfigs := make([][]byte, 0, len(configFiles))
		for _, file := range configFiles {
			if !strings.HasSuffix(file, ".json") {
				log.With("file", file).Error("multi-file CONFIG_FILES only supports .json overlays")
				os.Exit(1)
			}

			raw, readErr := os.ReadFile(file)
			if readErr != nil {
				log.With("error", readErr).Error("unable to read config from file (%s) error (%s) \n", file, readErr.Error())
				os.Exit(1)
			}

			if !isValidJSON(string(raw)) {
				log.Error("config file (%s) is not valid JSON", file)
				os.Exit(1)
			}

			rawConfigs = append(rawConfigs, raw)
		}

		merged := rawConfigs[0]
		for _, overlay := range rawConfigs[1:] {
			patchFromCompare, cmpErr := jsondiff.CompareJSON(merged, overlay)
			if cmpErr != nil {
				errors.Must(cmpErr)
			}

			// ignore removals, overlays are incremental and not complete
			var patch []jsondiff.Operation
			for idx := range patchFromCompare {
				if patchFromCompare[idx].Type == jsondiff.OperationRemove {
					continue
				}
				patch = append(patch, patchFromCompare[idx])
			}

			rawPatch, _ := gojson.Marshal(patch)
			patchOper, decErr := jsonpatch.DecodePatch(rawPatch)
			if decErr != nil {
				errors.Must(decErr)
			}

			next, applyErr := patchOper.Apply(merged)
			if applyErr != nil {
				errors.Must(applyErr)
			}

			merged = next
		}

		resolved, resolveErr := StructFromJSONBytesWithEnvReplace(merged, cfg, log)
		if resolveErr != nil {
			log.With("error", resolveErr).Error("unable to resolve merged config: %s", resolveErr.Error())
			os.Exit(1)
		}

		jsonBlob = resolved
		loaded = true
	}

	if !loaded {
		log.Error("configuration not loaded. see help.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Check if config validation is active from the environment
	checkConfig := internalEnv.IsConfigChecksActive()
	checkConfig = checkConfig && (len(withValidations) == 0 || withValidations[0])

	if checkConfig {
		missingFields, err := validateConfigTags(cfg, jsonBlob)
		if err != nil {
			log.
				With("missing_fields", missingFields).
				Error("configuration validation failed: %s", err)

			os.Exit(1)
		}
	}

	log.Debug("configuration loaded")
}

func validateConfigTags(cfg interface{}, jsonString []byte) ([]string, error) {
	expectedTags := collectJSONTags(reflect.ValueOf(cfg))

	var jsonMap map[string]gojson.RawMessage
	err := gojson.Unmarshal(jsonString, &jsonMap)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config JSON: %s", err.Error())
	}

	missingTags := []string{}
	for _, tag := range expectedTags {
		if !jsonContainsPath(jsonMap, strings.Split(tag, ".")) {
			missingTags = append(missingTags, tag)
		}
	}

	if len(missingTags) > 0 {
		return missingTags, errors.New("missing required config parameters")
	}

	return nil, nil
}

func collectJSONTags(v reflect.Value) []string {
	var tags []string
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	if t.Kind() != reflect.Struct {
		return tags
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		tag = strings.Split(tag, ",")[0]
		fieldValue := v.Field(i)

		if fieldValue.Kind() == reflect.Struct {
			nestedTags := collectJSONTags(fieldValue)
			for _, nestedTag := range nestedTags {
				tags = append(tags, tag+"."+nestedTag)
			}
		} else {
			tags = append(tags, tag)
		}
	}

	return tags
}

func jsonContainsPath(jsonMap map[string]gojson.RawMessage, path []string) bool {
	if len(path) == 0 {
		return true
	}

	value, ok := jsonMap[path[0]]
	if !ok {
		return false
	}

	if len(path) == 1 {
		return true
	}

	var nestedMap map[string]gojson.RawMessage
	err := gojson.Unmarshal(value, &nestedMap)
	if err != nil {
		return false
	}

	return jsonContainsPath(nestedMap, path[1:])
}
