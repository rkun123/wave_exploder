package songlink

import "context"

type contextKey struct{}

var songlinkKey = contextKey{}

type Songlink struct {
	base    string
	version string
}

// APIリクエストのボディ
type apiRequest struct {
}

func InitSonglink(ctx context.Context) context.Context {
	return context.WithValue(ctx, songlinkKey, newSonglink())
}

func newSonglink() *Songlink {
	return &Songlink{
		base:    "https://api.song.link",
		version: "v1-alpha.1",
	}
}

func GetSonglink(ctx context.Context) *Songlink {
	return ctx.Value(songlinkKey).(*Songlink)
}

func (s Songlink) baseURL() string {
	return s.base + "/" + s.version
}
