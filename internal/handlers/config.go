package handlers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/stormlightlabs/noteleaf/internal/store"
)

// ConfigHandler handles [store.Config]-related operations
type ConfigHandler struct {
	config *store.Config
}

// NewConfigHandler creates a new [ConfigHandler]
func NewConfigHandler() (*ConfigHandler, error) {
	config, err := store.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &ConfigHandler{config: config}, nil
}

// Get displays one or all configuration values
func (h *ConfigHandler) Get(key string) error {
	if key == "" {
		return h.displayAll()
	}

	value, err := h.getConfigValue(key)
	if err != nil {
		return err
	}

	fmt.Printf("%s = %v\n", key, value)
	return nil
}

// Set updates a configuration value
func (h *ConfigHandler) Set(key, value string) error {
	if err := h.setConfigValue(key, value); err != nil {
		return err
	}

	if err := store.SaveConfig(h.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}

// Path displays the configuration file path
func (h *ConfigHandler) Path() error {
	path, err := store.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	fmt.Println(path)
	return nil
}

// Reset resets the configuration to defaults
func (h *ConfigHandler) Reset() error {
	h.config = store.DefaultConfig()

	if err := store.SaveConfig(h.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Configuration reset to defaults")
	return nil
}

func (h *ConfigHandler) displayAll() error {
	v := reflect.ValueOf(*h.config)
	t := reflect.TypeOf(*h.config)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		tomlTag := field.Tag.Get("toml")
		if tomlTag == "" {
			continue
		}

		tagName := strings.Split(tomlTag, ",")[0]

		switch value.Kind() {
		case reflect.String:
			if value.String() != "" {
				fmt.Printf("%s = %q\n", tagName, value.String())
			} else {
				fmt.Printf("%s = \"\"\n", tagName)
			}
		case reflect.Bool:
			fmt.Printf("%s = %t\n", tagName, value.Bool())
		default:
			fmt.Printf("%s = %v\n", tagName, value.Interface())
		}
	}

	return nil
}

func (h *ConfigHandler) getConfigValue(key string) (interface{}, error) {
	v := reflect.ValueOf(*h.config)
	t := reflect.TypeOf(*h.config)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tomlTag := field.Tag.Get("toml")
		if tomlTag == "" {
			continue
		}

		tagName := strings.Split(tomlTag, ",")[0]
		if tagName == key {
			return v.Field(i).Interface(), nil
		}
	}

	return nil, fmt.Errorf("unknown config key: %s", key)
}

func (h *ConfigHandler) setConfigValue(key, value string) error {
	v := reflect.ValueOf(h.config).Elem()
	t := reflect.TypeOf(*h.config)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tomlTag := field.Tag.Get("toml")
		if tomlTag == "" {
			continue
		}

		tagName := strings.Split(tomlTag, ",")[0]
		if tagName == key {
			fieldValue := v.Field(i)

			switch fieldValue.Kind() {
			case reflect.String:
				fieldValue.SetString(value)
			case reflect.Bool:
				boolVal := value == "true" || value == "1" || value == "yes"
				fieldValue.SetBool(boolVal)
			default:
				return fmt.Errorf("unsupported field type for key %s", key)
			}

			return nil
		}
	}

	return fmt.Errorf("unknown config key: %s", key)
}
