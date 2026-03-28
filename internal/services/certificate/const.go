package certificate

// labels for tracking certificate created by this system
const (
	LabelManagedBy   = "app.kubernetes.io/managed-by"
	LabelOwnerSub    = "conduit.homelab.dev/owner-sub"
	LabelOwnerIss    = "conduit.homelab.dev/owner-iss"
	LabelRequestID   = "conduit.homelab.dev/request-id"
	ManagedByConduit = "conduit"
)
