package plugins

// Manifest declares a plugin's identity + contracts.
type Manifest struct {
	ID        string
	Version   string
	Contracts []string // capabilities this plugin registers
}

// Plugin is the single entrypoint a plugin exports.
type Plugin interface {
	Manifest() Manifest
	Register(api API) error
}

// API is the only surface plugins use to extend Mimir (the frozen contract - C2/ADR-005).
// Plugins import this package only; never core internals.
type API interface {
	RegisterProvider(def ProviderDef) error
	RegisterTool(def ToolDef) error
	RegisterHook(def HookDef) error
	RegisterChannel(def ChannelDef) error
	RegisterHarness(def HarnessDef) error
	RegisterCapability(def CapabilityDef) error
	RegisterCommand(def CommandDef) error
}

// Def types are intentionally minimal here; fleshed out per capability epic.
type ProviderDef struct{ ID, Dialect, BaseURL string }
type ToolDef struct{ Name, Description string }
type HookDef struct{ Event string }
type ChannelDef struct{ ID string }
type HarnessDef struct{ ID string }
type CapabilityDef struct{ Kind string }
type CommandDef struct{ Name string }

// Registry collects plugin registrations (one-way: plugins register, core reads).
type Registry struct {
	plugins   map[string]Manifest
	tools     []ToolDef
	providers []ProviderDef
	hooks     []HookDef
}

func NewRegistry() *Registry {
	return &Registry{plugins: map[string]Manifest{}}
}

// Load registers a plugin and records its manifest.
func (r *Registry) Load(p Plugin) error {
	m := p.Manifest()
	if err := p.Register(r); err != nil {
		return err
	}
	r.plugins[m.ID] = m
	return nil
}

func (r *Registry) RegisterProvider(d ProviderDef) error     { r.providers = append(r.providers, d); return nil }
func (r *Registry) RegisterTool(d ToolDef) error             { r.tools = append(r.tools, d); return nil }
func (r *Registry) RegisterHook(d HookDef) error             { r.hooks = append(r.hooks, d); return nil }
func (r *Registry) RegisterChannel(d ChannelDef) error       { return nil }
func (r *Registry) RegisterHarness(d HarnessDef) error       { return nil }
func (r *Registry) RegisterCapability(d CapabilityDef) error { return nil }
func (r *Registry) RegisterCommand(d CommandDef) error       { return nil }
