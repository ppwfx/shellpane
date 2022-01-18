package convert

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/ppwfx/shellpane/internal/bootstrap"
)

type SwaggerFile struct {
	Swagger string                          `json:"swagger"`
	Paths   map[string]map[string]Operation `json:"paths"`
}

type Parameters struct {
	Minimum     int    `json:"minimum,omitempty"`
	Type        string `json:"type"`
	Format      string `json:"format,omitempty"`
	XGoName     string `json:"x-go-name"`
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description,omitempty"`
	Ref         string `json:"$ref,omitempty"`
	Required    bool   `json:"required"`
}

type Operation struct {
	Description string              `json:"description"`
	Produces    []string            `json:"produces"`
	Summary     string              `json:"summary"`
	OperationID string              `json:"operationId"`
	Parameters  []Parameters        `json:"parameters"`
	Responses   map[string]Response `json:"responses"`
}

type Response struct {
	Description string `json:"description"`
}

func swaggerParameterNameToInputSlug(n string) string {
	return strings.ToUpper(n)
}

var matchNonAlphaNum = regexp.MustCompile(`[^a-zA-Z0-9]`)
var matchTwoOrMoreDashes = regexp.MustCompile(`-{2,}`)

func swaggerOperationToCommandSlug(path string, method string) string {
	slug := fmt.Sprintf("%v-%v", method, path)
	slug = matchNonAlphaNum.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	slug = matchTwoOrMoreDashes.ReplaceAllString(slug, "-")
	slug = strings.ToLower(slug)

	return slug
}

func ConvertSwaggerfileToShellpaneConfig(f SwaggerFile, category string) (bootstrap.ShellpaneConfig, error) {
	var inputConfigs []bootstrap.InputConfig
	seenInputs := map[string]struct{}{}
	for path, _ := range f.Paths {
		for _, operation := range f.Paths[path] {
			for _, parameter := range operation.Parameters {
				inputSlug := swaggerParameterNameToInputSlug(parameter.Name)
				_, ok := seenInputs[inputSlug]
				if ok {
					continue
				}
				seenInputs[inputSlug] = struct{}{}

				inputConfigs = append(inputConfigs, bootstrap.InputConfig{
					Slug:        inputSlug,
					Description: parameter.Description,
				})
			}
		}
	}

	var commandConfigs []bootstrap.CommandConfig
	for path, _ := range f.Paths {
		for method, operation := range f.Paths[path] {
			slug := swaggerOperationToCommandSlug(path, method)

			var commandInputConfigs []bootstrap.CommandInputConfig
			for _, parameter := range operation.Parameters {
				commandInputConfigs = append(commandInputConfigs, bootstrap.CommandInputConfig{
					InputSlug: swaggerParameterNameToInputSlug(parameter.Name),
				})
			}

			u, err := url.Parse(path)
			if err != nil {
				return bootstrap.ShellpaneConfig{}, errors.Wrapf(err, "failed to url parse path=%v", path)
			}

			q := "?"
			for _, parameter := range operation.Parameters {
				switch parameter.In {
				case "query":
					q = fmt.Sprintf("%v%v=$%v&", q, parameter.Name, swaggerParameterNameToInputSlug(parameter.Name))
				case "path":
					u.Path = strings.Replace(u.Path, fmt.Sprintf("{%v}", parameter.Name), fmt.Sprintf("$%v", swaggerParameterNameToInputSlug(parameter.Name)), -1)
				}
			}

			path := strings.TrimRight(fmt.Sprintf(`%v%v`, u.Path, q), "&?")

			commandConfigs = append(commandConfigs, bootstrap.CommandConfig{
				Slug:        slug,
				Command:     fmt.Sprintf(`http %v "$HOST%v" | jq .`, method, path),
				Inputs:      commandInputConfigs,
				Description: operation.Description,
			})
		}
	}

	var viewConfigs []bootstrap.ViewConfig
	for path, _ := range f.Paths {
		for method, _ := range f.Paths[path] {
			slug := swaggerOperationToCommandSlug(path, method)

			viewConfigs = append(viewConfigs, bootstrap.ViewConfig{
				Slug:         slug,
				CategorySlug: category,
				Name:         strings.TrimSpace(fmt.Sprintf("%v %v", method, path)),
				CommandSlug:  slug,
			})
		}
	}

	c := bootstrap.ShellpaneConfig{
		Categories: []bootstrap.CategoryConfig{
			{
				Slug:  category,
				Name:  category,
				Color: "#ff0374",
			},
		},
		Views:    viewConfigs,
		Commands: commandConfigs,
		Inputs:   inputConfigs,
	}

	return c, nil
}
