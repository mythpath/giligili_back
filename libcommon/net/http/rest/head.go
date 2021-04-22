package rest

const GiligiliUser = "user"
const GiligiliToken = "gtoken"

// RestContextKey
type RestContextKey string

// CurrentUserKey returns current user key in the golang context
func CurrentUserKey() RestContextKey {
	return RestContextKey(GiligiliUser)
}

// CurrentUserKey returns current auth info in the golang context
func CurrentAuthInfoKey() RestContextKey {
	return RestContextKey("passport-auth")
}

func CurrentTenantKey() interface{} {
	return RestContextKey("passport-tenant")
}
