package types

type Device struct {
	ID             uint
	Manufacturer   string
	Model          string
	OS             string
	Build          string
	CPUDescription string
	DisplayX       uint
	DisplayY       uint
	AndroidVersion string
	DPI            uint
	BuildOSDevice  string
}
