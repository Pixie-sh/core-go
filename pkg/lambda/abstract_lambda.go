package lambda

import "context"

type AbstractLambda struct {
	config AbstractLambdaConfiguration
}

func (gl AbstractLambda) Init(ctx context.Context) error {
	return initLambda(ctx)
}

type AbstractLambdaConfiguration struct {
	AppVersion string
}

func (c AbstractLambdaConfiguration) Load(ctx context.Context) error {
	return nil //nothing for now
}
