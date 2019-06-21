package component

import (
	"fmt"

	templatev1 "github.com/openshift/api/template/v1"
)

type Productized struct {
	options []string
	Options *ProductizedOptions
}

func NewProductized(options []string) *Productized {
	productized := &Productized{
		options: options,
	}
	return productized
}

type ProductizedOptions struct {
	productizedNonRequiredOptions
	productizedRequiredOptions
}

type productizedRequiredOptions struct {
	ampRelease   string
	apicastImage string
	backendImage string
	systemImage  string
	zyncImage    string
}

type productizedNonRequiredOptions struct {
}

type ProductizedOptionsProvider interface {
	GetProductizedOptions() *ProductizedOptions
}
type CLIProductizedOptionsProvider struct {
}

func (o *CLIProductizedOptionsProvider) GetProductizedOptions() (*ProductizedOptions, error) {
	pob := ProductizedOptionsBuilder{}
	pob.ApicastImage("${AMP_APICAST_IMAGE}")
	pob.BackendImage("${AMP_BACKEND_IMAGE}")
	pob.SystemImage("${AMP_SYSTEM_IMAGE}")
	pob.ZyncImage("${AMP_ZYNC_IMAGE}")
	res, err := pob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Productized Options - %s", err)
	}
	return res, nil
}

func (productized *Productized) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
}

func (productized *Productized) PostProcess(template *templatev1.Template, otherComponents []Component) {
	// TODO move this outside this specific method
	optionsProvider := CLIProductizedOptionsProvider{}
	productizedOpts, err := optionsProvider.GetProductizedOptions()
	_ = err
	productized.Options = productizedOpts
	res := template.Objects
	productized.updateAmpImagesParameters(template)
	template.Objects = res
}

func (productized *Productized) updateAmpImagesParameters(template *templatev1.Template) {
	for paramIdx := range template.Parameters {
		param := &template.Parameters[paramIdx]
		switch param.Name {
		case "AMP_SYSTEM_IMAGE":
			param.Value = "registry.redhat.io/3scale-amp26/system"
		case "AMP_BACKEND_IMAGE":
			param.Value = "registry.redhat.io/3scale-amp26/backend"
		case "AMP_APICAST_IMAGE":
			param.Value = "registry.redhat.io/3scale-amp26/apicast-gateway"
		case "AMP_ZYNC_IMAGE":
			param.Value = "registry.redhat.io/3scale-amp26/zync"
		}
	}
}
