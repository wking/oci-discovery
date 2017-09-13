package discovery

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/xiekeyang/oci-discovery/tools/object"
)

var (
	defaultTrans = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
)

func DiscoveryHandler(context *cli.Context) error {
	var (
		name  = context.Args()[0]
		roots []v1.Descriptor
	)

	v, err := paramsParser(name)
	if err != nil {
		return err
	}

	engines, err := refEnginesFetching(v)
	if err != nil {
		return err
	}

	for _, engine := range engines.RefEngines {
		var ur urlResolver = urlResolver(engine.Uri)
		u, err := ur.resolve(v)
		if err != nil {
			return err
		}

		index, err := ociIndexFetching(u)
		if err != nil {
			return err
		}

		if fragment, ok := v["fragment"]; ok {
			for _, manifest := range index.Manifests {
				if fragment == manifest.Annotations[`org.opencontainers.image.ref.name`] {
					roots = append(roots, manifest)
				}
			}
		} else {
			roots = append(roots, index.Manifests...)
		}
	}

	return stdWrite(roots)
}

func refEnginesFetching(v map[string]interface{}) (*object.RefEngines, error) {
	u, err := templateRefEngines.resolve(v)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Transport: defaultTrans}

	resp, err := client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ref engine fetching error, status code = %d", resp.StatusCode)
	}

	var engines object.RefEngines
	if err := json.NewDecoder(resp.Body).Decode(&engines); err != nil {
		logrus.Errorf("ref engines object decoded failed: %s", err)
		return nil, err
	}

	return &engines, nil
}

func ociIndexFetching(u *url.URL) (*v1.Index, error) {
	var index *v1.Index

	client := &http.Client{Transport: defaultTrans}

	resp, err := client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		logrus.Errorf("index decoded failed: %s", err)
		return nil, err
	}

	return index, nil
}

func stdWrite(v interface{}) error {
	var out bytes.Buffer

	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	err = json.Indent(&out, b, "", "\t")
	if err != nil {
		return err
	}

	_, err = out.WriteTo(os.Stdout)
	if err != nil {
		return err
	}

	return nil
}
