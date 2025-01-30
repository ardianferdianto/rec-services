package contextprop

import "context"

type ContextKey string

const (
	CorrelationIDKey ContextKey = "CorrelationID"
	ClientIDKey      ContextKey = "ClientID"
	LocalizerKey     ContextKey = "Localizer"
	UserDataKey      ContextKey = "UserData"
)

func GetValue(ctx context.Context, key ContextKey) string {
	value, ok := ctx.Value(key).(string)
	if !ok {
		return ""
	}
	return value
}

func SetValue(ctx context.Context, key ContextKey, value string) context.Context {
	return context.WithValue(ctx, key, value)
}
