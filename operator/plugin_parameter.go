package operator

import (
	"fmt"

	"github.com/observiq/stanza/errors"
)

// PluginParameter is a basic description of a plugin's parameter.
type PluginParameter struct {
	Label       string
	Description string
	Required    bool
	Type        string      // "string", "int", "bool", "strings", or "enum"
	ValidValues []string    `yaml:"valid_values"` // only useable if Type == "enum"
	Default     interface{} // Must be valid according to Type & ValidValues
}

func (param PluginParameter) validate() error {
	if param.Required && param.Default != nil {
		return errors.NewError(
			"required parameter cannot have a default value",
			"ensure that required parameters do not have default values",
		)
	}

	if err := param.validateType(); err != nil {
		return err
	}

	if err := param.validateValidValues(); err != nil {
		return err
	}

	if err := param.validateDefault(); err != nil {
		return err
	}

	return nil
}

func (param PluginParameter) validateType() error {
	switch param.Type {
	case "string", "int", "bool", "strings", "enum": // ok
	default:
		return errors.NewError(
			"invalid type for parameter",
			"ensure that the type is one of 'string', 'int', 'bool', 'strings', or 'enum'",
		)
	}
	return nil
}

func (param PluginParameter) validateValidValues() error {
	switch param.Type {
	case "string", "int", "bool", "strings":
		if len(param.ValidValues) > 0 {
			return errors.NewError(
				fmt.Sprintf("valid_values is undefined for parameter of type '%s'", param.Type),
				"remove 'valid_values' field or change type to 'enum'",
			)
		}
	case "enum":
		if len(param.ValidValues) == 0 {
			return errors.NewError(
				"parameter of type 'enum' must have 'valid_values' specified",
				"specify an array that includes one or more valid values",
			)
		}
	}
	return nil
}

func (param PluginParameter) validateDefault() error {
	if param.Default == nil {
		return nil
	}

	// Validate that Default corresponds to Type
	switch param.Type {
	case "string":
		return validateStringDefault(param)
	case "int":
		return validateIntDefault(param)
	case "bool":
		return validateBoolDefault(param)
	case "strings":
		return validateStringArrayDefault(param)
	case "enum":
		return validateEnumDefault(param)
	default:
		return errors.NewError(
			"invalid type for parameter",
			"ensure that the type is one of 'string', 'int', 'bool', 'strings', or 'enum'",
		)
	}
}

func validateStringDefault(param PluginParameter) error {
	if _, ok := param.Default.(string); !ok {
		return errors.NewError(
			"default value for a parameter of type 'string' must be a string",
			"ensure that the default value is a string",
		)
	}
	return nil
}

func validateIntDefault(param PluginParameter) error {
	switch param.Default.(type) {
	case int, int32, int64:
		return nil
	default:
		return errors.NewError(
			"default value for a parameter of type 'int' must be an integer",
			"ensure that the default value is an integer",
		)
	}
}

func validateBoolDefault(param PluginParameter) error {
	if _, ok := param.Default.(bool); !ok {
		return errors.NewError(
			"default value for a parameter of type 'bool' must be a boolean",
			"ensure that the default value is a boolean",
		)
	}
	return nil
}

func validateStringArrayDefault(param PluginParameter) error {
	defaultList, ok := param.Default.([]interface{})
	if !ok {
		return errors.NewError(
			"default value for a parameter of type 'strings' must be an array of strings",
			"ensure that the default value is a string",
		)
	}
	for _, s := range defaultList {
		if _, ok := s.(string); !ok {
			return errors.NewError(
				"default value for a parameter of type 'strings' must be an array of strings",
				"ensure that the default value is an array of strings",
			)
		}
	}
	return nil
}

func validateEnumDefault(param PluginParameter) error {
	def, ok := param.Default.(string)
	if !ok {
		return errors.NewError(
			"invalid default for enumerated parameter",
			"ensure that the default value is a string",
		)
	}
	for _, val := range param.ValidValues {
		if val == def {
			return nil
		}
	}
	return errors.NewError(
		"invalid default value for enumerated parameter",
		"ensure default value is listed as a valid value",
	)
}
