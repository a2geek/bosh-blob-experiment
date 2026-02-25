package manifest

type Job struct {
	Name        string
	Version     string
	Fingerprint string
	Sha1        string
	Packages    []string
}

type Package struct {
	Name         string
	Version      string
	Fingerprint  string
	Sha1         string
	Dependencies []string
}

type ReleaseManifest struct {
	Name              string
	Version           string
	CommitHash        string
	UncommitedChanges bool
	Jobs              []Job
	Packages          []Package
}
