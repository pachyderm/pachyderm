package jobs

// Reference is something that can match a build input or artifact.
type Reference interface{ Match(any) bool }
type WithName interface{ GetName() string }
type WithPlatform interface{ GetPlatform() Platform }
type ConstrainableToPlatform interface{ ConstrainToPlatform(Platform) Reference }

// Name references something by name.
type Name string

var _ Reference = Name("")
var _ WithName = Name("")
var _ ConstrainableToPlatform = Name("")

// NameAndPlatform references something by name and platform.
type NameAndPlatform struct {
	Name     string
	Platform Platform
}

var _ Reference = (*NameAndPlatform)(nil)
var _ WithName = (*NameAndPlatform)(nil)
var _ WithPlatform = (*NameAndPlatform)(nil)

func (me Name) GetName() string { return string(me) }

func (me Name) Match(target any) bool {
	if x, ok := target.(WithName); ok {
		return me.GetName() == x.GetName()
	}
	return me == target
}

func (me Name) ConstrainToPlatform(p Platform) Reference {
	return NameAndPlatform{
		Name:     me.GetName(),
		Platform: p,
	}
}

func (me Name) String() string {
	return me.GetName()
}

func (me NameAndPlatform) GetName() string       { return me.Name }
func (me NameAndPlatform) GetPlatform() Platform { return me.Platform }
func (me NameAndPlatform) Match(target any) bool {
	if me == target {
		return true
	}
	if me.GetPlatform() == AllPlatforms {
		return true
	}
	if x, ok := target.(WithName); !ok || me.GetName() != x.GetName() {
		return false
	}
	if x, ok := target.(WithPlatform); !ok || (me.GetPlatform() != x.GetPlatform()) {
		return false
	}
	return true
}

func (me NameAndPlatform) String() string {
	return me.Name + "#" + string(me.Platform)
}
