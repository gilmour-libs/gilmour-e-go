package backends

import (
	"os"
	"testing"

	"gopkg.in/gilmour-libs/gilmour-e-go.v5/backends/redis"
)

var backend Backend

func makeBackend() Backend {
	return redis.MakeRedis("localhost", "")
}

func TestSupportedPolicies(t *testing.T) {
	can_be := []string{
		ErrorPolicyQueue,
		ErrorPolicyPublish,
		ErrorPolicyIgnore,
	}

	policies := backend.SupportedErrorPolicies()
	for _, p := range policies {
		found := false
		for _, cp := range can_be {
			if cp == p {
				found = true
				break
			}
		}

		if !found {
			t.Error("Policy must be one of those supported by Gilmour")
		}
	}
}

func TestGetErrorPolicyDefault(t *testing.T) {
	// Should return a default Policy, even when one is not set.
}

func TestSetErrorPolicyFail(t *testing.T) {
	// Should raise an error if the policy is not one supported by Gilmour
}

func TestSetErrorPolicy(t *testing.T) {
	//Should set and get the appropriate error policy.
}

func TestHasActiveSubsribers(t *testing.T) {
	//Should return false for an unsubscribed topic.
}

func TestIsActiveSubsribed(t *testing.T) {
	//Should return false for an unsubscribed topic.
}

func TestReportError(t *testing.T) {
	//Set policy as queue and Should queue errors
}

func TestSubscribe(t *testing.T) {
	//Should subscribe a topic, IsTopicSubscribed should be true.
	//Should unsubscribe, IsTopicSubscribed should be false.
}

func TestPublish(t *testing.T) {
	//Should return num 0 for an unsubscribed topic.
}

func TestActiveIdents(t *testing.T) {
	//Should return a map[string]string
}

func TestRegisterIdent(t *testing.T) {
	//Should set an ident and be a part of ActiveIdents()
	//Should unregister the ident and not be a part of ActiveIdents()
}

func TestPacketPub(t *testing.T) {
	// Start should return a channel, which must emit packet on the channel
	// for a subscribed channel.
}

func TestMain(m *testing.M) {
	backend = makeBackend()
	status := m.Run()
	os.Exit(status)
}
