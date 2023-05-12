package bff

type CandidateUser struct {
	IsArtist bool
	// just for internal use
	Comment string
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
			Comment:  "Noah (ottr.sh)",
			IsArtist: false,
		},
		"did:plc:jt43524ltn23seg5v3qhurwt": {
			Comment:  "vilk (vilk.pub)",
			IsArtist: false,
		},
		"did:plc:ouytv644apqbu2pm7fnp7qrj": {
			Comment:  "Newton (newton.dog)",
			IsArtist: true,
		},
		"did:plc:hjzrjs7sewv6nmratpoeavtp": {
			Comment:  "Kepler",
			IsArtist: false,
		},
		"did:plc:ojw5gcvjs44m7dl5zrzeb4i3": {
			Comment:  "Rend (dingo.bsky.social)",
			IsArtist: false,
		},
		"did:plc:ggg7g6gcc65lzwqvqqpa2mik": {
			Comment:  "Concoction (concoction.bsky.social)",
			IsArtist: true,
		},
		"did:plc:sfvpv6dfrug3rnjewn7gyx62": {
			Comment:  "qdot (buttplug.engineer)",
			IsArtist: false,
		},
		"did:plc:o74zbazekchwk2v4twee4ekb": {
			Comment:  "kio (kio.dev)",
			IsArtist: false,
		},
		"did:plc:rgbf6ph3eki5lffvrs6syf4w": {
			Comment:  "cael (cael.tech)",
			IsArtist: false,
		},
		"did:plc:wtfep3izymr6ot4tywoqcydc": {
			Comment:  "adam (snowfox.gay)",
			IsArtist: false,
		},
		"did:plc:6aikzgasri74fypm4h3qfvui": {
			Comment:  "havokhusky (havok.bark.supply)",
			IsArtist: false,
		},
		"did:plc:rjawzv3m7smnyaiq62mrqpok": {
			Comment:  "frank (lickmypa.ws)",
			IsArtist: false,
		},
		"did:plc:f3ynrkwdfe7m5ffvxd5pxf4f": {
			Comment:  "lobo (lupine.agency)",
			IsArtist: false,
		},
	}
}
