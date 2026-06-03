package flickr

import (
	"context"
	"encoding/json"
)

// GetMethods calls flickr.reflection.getMethods and returns the raw result.
func (c *Client) GetMethods(ctx context.Context) (json.RawMessage, error) {
	return c.CallRaw(ctx, "flickr.reflection.getMethods", nil)
}

// GetMethodInfo calls flickr.reflection.getMethodInfo for the given method name.
func (c *Client) GetMethodInfo(ctx context.Context, methodName string) (json.RawMessage, error) {
	return c.CallRaw(ctx, "flickr.reflection.getMethodInfo", map[string]string{
		"method_name": methodName,
	})
}
