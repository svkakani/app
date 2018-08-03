package dockerapp


// DockerAppRaw is the raw representation of an application
type DockerAppRaw struct {
	Metadata     []byte
	Compose      []byte
	Settings     []byte
}

// DockerApp is the main structure that represents a docker application
type DockerApp struct {
	AppName   string // The application name without extension
	Origin    string // The input application path or image if relevant
	Raw       *DockerAppRaw // The raw application content
}