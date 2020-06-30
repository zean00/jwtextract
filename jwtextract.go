package jwtextract

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"
)

const authHeader = "Authorization"
const Namespace = "github.com/zean00/jwtextract"

// ProxyFactory creates an proxy factory over the injected one adding a JSON Schema
// validator middleware to the pipe when required
func ProxyFactory(pf proxy.Factory) proxy.FactoryFunc {
	return proxy.FactoryFunc(func(cfg *config.EndpointConfig) (proxy.Proxy, error) {
		next, err := pf.New(cfg)
		if err != nil {
			return next, err
		}
		claimMap, ok := configGetter(cfg.ExtraConfig).(map[string]string)
		if !ok {
			return next, nil
		}
		return newProxy(claimMap, next), nil
	})
}

func newProxy(claimMap map[string]string, next proxy.Proxy) proxy.Proxy {
	return func(ctx context.Context, r *proxy.Request) (*proxy.Response, error) {
		if err := extractClaim(claimMap, r); err != nil {
			fmt.Println(err)
			return next(ctx, r)
		}
		return next(ctx, r)
	}
}

func configGetter(cfg config.ExtraConfig) interface{} {
	v, ok := cfg[Namespace]
	if !ok {
		return nil
	}
	return v
}

func extractClaim(claimMap map[string]string, r *proxy.Request) error {
	token := r.Headers[authHeader][0]
	if token == "" {
		return errors.New("Token is empty, skip extracting ")
	}
	token = strings.TrimPrefix(token, "Bearer ")
	//Just in case using lower case bearer
	token = strings.TrimPrefix(token, "bearer ")
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return errors.New("Token is malformed , skip processing")
	}

	payload := parts[1]
	claim, err := base64.RawStdEncoding.DecodeString(payload)
	if err != nil {
		return errors.New("Unable to decode payload, " + err.Error())
	}

	var data map[string]interface{}
	if err := json.Unmarshal(claim, &data); err != nil {
		return errors.New("Unable to unmarshal claim " + err.Error())
	}

	for k, v := range data {
		key, ok := claimMap[k]
		if ok {
			r.Headers[key] = []string{fmt.Sprintf("%v", v)}
		} else {
			r.Headers[k] = []string{fmt.Sprintf("%v", v)}
		}
	}
	return nil
}
