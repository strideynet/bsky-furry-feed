package main

type CandidateUser struct {
	IsArtist bool
	// just for internal use
	comment string
}

type StaticCandidateUsers map[string]CandidateUser

func (c StaticCandidateUsers) GetByDID(did string) *CandidateUser {
	v, ok := c[did]
	if ok {
		return &v
	}
	return nil
}

func NewStaticCandidateUsers() StaticCandidateUsers {
	return StaticCandidateUsers{
		"did:plc:dllwm3fafh66ktjofzxhylwk": {
			comment:  "Noah (ottr.sh)",
			IsArtist: false,
		},
		"did:plc:jt43524ltn23seg5v3qhurwt": {
			comment:  "vilk (vilk.pub)",
			IsArtist: false,
		},
		"did:plc:ouytv644apqbu2pm7fnp7qrj": {
			comment:  "Newton (newton.dog)",
			IsArtist: true,
		},
		"did:plc:hjzrjs7sewv6nmratpoeavtp": {
			comment:  "Kepler",
			IsArtist: false,
		},
		"did:plc:ojw5gcvjs44m7dl5zrzeb4i3": {
			comment:  "Rend (dingo.bsky.social)",
			IsArtist: false,
		},
		"did:plc:ggg7g6gcc65lzwqvqqpa2mik": {
			comment:  "Concoction (concoction.bsky.social)",
			IsArtist: true,
		},
		"did:plc:sfvpv6dfrug3rnjewn7gyx62": {
			comment:  "qdot (buttplug.engineer)",
			IsArtist: false,
		},
		"did:plc:o74zbazekchwk2v4twee4ekb": {
			comment:  "kio (kio.dev)",
			IsArtist: false,
		},
		"did:plc:rgbf6ph3eki5lffvrs6syf4w": {
			comment:  "cael (cael.tech)",
			IsArtist: false,
		},
		"did:plc:wtfep3izymr6ot4tywoqcydc": {
			comment:  "adam (snowfox.gay)",
			IsArtist: false,
		},
	}
}
