package normalize

import "github.com/wazp/c64dreams-tool/pkg/model"

// Options control normalization behavior.
type Options struct {
	Target     model.TargetDevice
	MaxNameLen int
}

// EffectiveMaxLen resolves the maximum name length using overrides or target defaults.
func (o Options) EffectiveMaxLen() int {
	if o.MaxNameLen > 0 {
		return o.MaxNameLen
	}
	return model.ProfileFor(o.Target).MaxNameLen
}
