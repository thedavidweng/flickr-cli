package config

// ProfileNames returns the sorted list of profile names in the config.
func (c *Config) ProfileNames() []string {
	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}
	return names
}

// ActiveProfile returns the current profile name, falling back to "default".
func (c *Config) ActiveProfile() string {
	if c.CurrentProfile != "" {
		return c.CurrentProfile
	}
	return "default"
}
