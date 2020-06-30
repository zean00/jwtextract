package jwtextract

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/proxy"
)

const authHeader = "Authorization"
const Namespace = "github_com/zean00/jwtextract"

type xtraConfig struct {
	ExtractAll bool
	ClaimMap   map[string]interface{}
}

// ProxyFactory creates an proxy factory over the injected one adding a JSON Schema
// validator middleware to the pipe when required
func ProxyFactory(l logging.Logger, pf proxy.Factory) proxy.FactoryFunc {
	return proxy.FactoryFunc(func(cfg *config.EndpointConfig) (proxy.Proxy, error) {
		next, err := pf.New(cfg)
		if err != nil {
			return next, err
		}

		conf := configGetter(cfg.ExtraConfig)

		if conf == nil {
			l.Debug("[jwtextract] No config for jwtextract ")
			return next, nil
		}

		l.Debug("[jwtextract] Claim map ", conf.ClaimMap)
		return newProxy(l, conf, next), nil
	})
}

func newProxy(l logging.Logger, config *xtraConfig, next proxy.Proxy) proxy.Proxy {
	return func(ctx context.Context, r *proxy.Request) (*proxy.Response, error) {
		if err := extractClaim(config, r); err != nil {
			l.Error("[jwtextract]", err)
			return next(ctx, r)
		}
		return next(ctx, r)
	}
}

func configGetter(cfg config.ExtraConfig) *xtraConfig {
	v, ok := cfg[Namespace]
	if !ok {
		return nil
	}
	tmp, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	conf := xtraConfig{
		ExtractAll: false,
		ClaimMap:   make(map[string]interface{}),
	}
	xa, ok := tmp["extract_all"].(bool)
	if ok {
		conf.ExtractAll = xa
	}
	cmap, ok := tmp["claim_map"].(map[string]interface{})
	if ok {
		conf.ClaimMap = cmap
	}
	return &conf
}

func extractClaim(config *xtraConfig, r *proxy.Request) error {
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
		key, ok := config.ClaimMap[k]
		if ok {
			r.Headers[key.(string)] = []string{fmt.Sprintf("%v", v)}
		} else {
			if config.ExtractAll {
				r.Headers[k] = []string{fmt.Sprintf("%v", v)}
			}
		}
	}
	return nil
}
