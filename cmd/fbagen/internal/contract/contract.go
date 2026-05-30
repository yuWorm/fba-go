package contract

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Contracts struct {
	API      APIContract
	Response ResponseContract
	Redis    RedisContract
}

type APIContract struct {
	Version        int     `yaml:"version"`
	BasePath       string  `yaml:"base_path"`
	PriorityRoutes []Route `yaml:"priority_routes"`
	Routes         []Route `yaml:"routes"`
}

type Route struct {
	Method           string         `yaml:"method" json:"method"`
	Path             string         `yaml:"path" json:"path"`
	SamplePath       string         `yaml:"sample_path,omitempty" json:"sample_path,omitempty"`
	Request          *RequestSample `yaml:"request,omitempty" json:"request,omitempty"`
	Owner            string         `yaml:"owner,omitempty" json:"owner,omitempty"`
	Permission       string         `yaml:"permission,omitempty" json:"permission,omitempty"`
	ResponseEnvelope *bool          `yaml:"response_envelope,omitempty" json:"response_envelope,omitempty"`
}

type RequestSample struct {
	ContentType string            `yaml:"content_type,omitempty" json:"content_type,omitempty"`
	Headers     map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Body        string            `yaml:"body,omitempty" json:"body,omitempty"`
}

type ResponseContract struct {
	Version    int             `yaml:"version"`
	Success    ResponseSuccess `yaml:"success"`
	Error      ResponseError   `yaml:"error"`
	Pagination Pagination      `yaml:"pagination"`
	Datetime   Datetime        `yaml:"datetime"`
}

type ResponseSuccess struct {
	Envelope       bool     `yaml:"envelope"`
	RequiredFields []string `yaml:"required_fields"`
	Code           int      `yaml:"code"`
	Msg            string   `yaml:"msg"`
}

type ResponseError struct {
	Envelope       bool     `yaml:"envelope"`
	RequiredFields []string `yaml:"required_fields"`
}

type Pagination struct {
	RequiredFields      []string `yaml:"required_fields"`
	LinksRequiredFields []string `yaml:"links_required_fields"`
}

type Datetime struct {
	Layout          string `yaml:"layout"`
	TimezoneDefault string `yaml:"timezone_default"`
}

type RedisContract struct {
	Version       int                 `yaml:"version"`
	DefaultPrefix string              `yaml:"default_prefix"`
	Keys          map[string]RedisKey `yaml:"keys"`
}

type RedisKey struct {
	Pattern string `yaml:"pattern"`
	TTL     string `yaml:"ttl,omitempty"`
	Type    string `yaml:"type,omitempty"`
	Note    string `yaml:"note,omitempty"`
}

func Load(dir string) (Contracts, error) {
	var contracts Contracts
	if err := readYAML(filepath.Join(dir, "api.contract.yaml"), &contracts.API); err != nil {
		return Contracts{}, err
	}
	if err := readYAML(filepath.Join(dir, "response.contract.yaml"), &contracts.Response); err != nil {
		return Contracts{}, err
	}
	if err := readYAML(filepath.Join(dir, "redis.contract.yaml"), &contracts.Redis); err != nil {
		return Contracts{}, err
	}
	return contracts, nil
}

func readYAML(path string, out any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(content, out)
}
