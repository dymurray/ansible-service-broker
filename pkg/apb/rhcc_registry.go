package apb

import (
	b64 "encoding/base64"
	"encoding/json"
	logging "github.com/op/go-logging"
	"io/ioutil"
	"net/http"
	"strings"
)

// RHCCRegistry - Red Hat Container Catalog Registry
type RHCCRegistry struct {
	config RegistryConfig
	log    *logging.Logger
}

type Image struct {
	Description  string `json:"description"`
	IsOfficial   bool   `json:"is_official"`
	IsTrusted    bool   `json:"is_trusted"`
	Name         string `json:"name"`
	ShouldFilter bool   `json:"should_filter"`
	StarCount    int    `json:"star_count"`
}

type ImageResponse struct {
	NumResults int      `json:"num_results"`
	Query      string   `json:"query"`
	Results    []*Image `json:"results"`
}

// Init - Initialize the Red Hat Container Catalog
func (r *RHCCRegistry) Init(config RegistryConfig, log *logging.Logger) error {
	log.Debug("RHCCRegistry::Init")
	r.config = config
	r.log = log
	return nil
}

// This function is used because our code expects an HTTP Url for talking to RHCC
func (r RHCCRegistry) cleanHttpUrl(url string) string {
	if strings.HasPrefix(url, "http://") == true {
		return url
	}

	if strings.HasPrefix(url, "https://") == true {
		return url
	}

	url = "http://" + url
	return url
}

// LoadSpecs - Load Red Hat Container Catalog specs
func (r RHCCRegistry) LoadSpecs() ([]*Spec, int, error) {
	r.log.Debug("RHCCRegistry::LoadSpecs")
	var specs []*Spec

	imageList, err := r.LoadImages("\"*-apb\"")
	if err != nil {
		return []*Spec{}, 0, err
	}

	numResults := imageList.NumResults
	r.log.Debug("Found %d images in RHCC", numResults)
	for _, image := range imageList.Results {
		if spec := r.imageToSpec(image); spec != nil {
			specs = append(specs, spec)
		}
	}

	return specs, numResults, nil
}

func (r RHCCRegistry) imageToSpec(image *Image) *Spec {
	r.log.Debug("RHCCRegistry::imageToSpec")
	_spec := &Spec{}
	url := r.cleanHttpUrl(r.config.Url)

	req, err := http.NewRequest("GET", url+"/v2/"+image.Name+"/manifests/latest", nil)
	if err != nil {
		return nil
	}

	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	type label struct {
		Spec    string `json:"com.redhat.apb.spec"`
		Version string `json:"com.redhat.apb.version"`
	}

	type config struct {
		Label label `json:"Labels"`
	}

	hist := struct {
		History []map[string]string `json:"history"`
	}{}

	conf := struct {
		Config *config `json:"config"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&hist)
	if err != nil {
		r.log.Info("Error grabbing JSON body from response: %s. Skipping.", err)
		return nil
	}

	if hist.History == nil {
		r.log.Info("V1 Schema Manifest history does not exist in registry. Skipping.")
		return nil
	}

	err = json.Unmarshal([]byte(hist.History[0]["v1Compatibility"]), &conf)
	if err != nil {
		r.log.Info("Error unmarshalling intermediary JSON response: %s. Skipping.", err)
		return nil
	}

	if conf.Config == nil {
		r.log.Info("Did not find v1 Manifest in image history. Skipping.")
		return nil
	}

	encodedSpec := conf.Config.Label.Spec
	if len(encodedSpec) == 0 {
		r.log.Info("Didn't find encoded Spec label. Assuming image is not APB and skipping.")
		return nil
	}

	decodedSpecYaml, err := b64.StdEncoding.DecodeString(encodedSpec)
	if err != nil {
		r.log.Info("Something went wrong decoding spec from label. Skipping.")
		return nil
	}

	if err = LoadYAML(string(decodedSpecYaml), _spec); err != nil {
		r.log.Info("Something went wrong loading decoded spec yaml, %s. Skipping.", err)
		return nil
	}
	r.log.Debug("Successfully converted RHCC Image %s into Spec", _spec.Name)

	return _spec
}

func (r RHCCRegistry) LoadImages(Query string) (ImageResponse, error) {
	r.log.Debug("RHCCRegistry::LoadImages")
	url := r.cleanHttpUrl(r.config.Url)
	r.log.Debug("Using " + url + " to source APB images using query:" + Query)
	req, err := http.NewRequest("GET", url+"/v1/search?q="+Query, nil)
	if err != nil {
		return ImageResponse{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ImageResponse{}, err
	}
	defer resp.Body.Close()

	r.log.Debug("Got Image Response from RHCC")
	imageList, err := ioutil.ReadAll(resp.Body)

	imageResp := ImageResponse{}
	err = json.Unmarshal(imageList, &imageResp)
	if err != nil {
		return ImageResponse{}, err
	}
	r.log.Debug("Properly unmarshalled image response")

	return imageResp, nil
}
