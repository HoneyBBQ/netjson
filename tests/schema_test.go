package netjson_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	pb "netjson/gen/go"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestNetJSONSamples(t *testing.T) {
	runRoundTrip(t, "openwrt", "openwrt/*.json", func() protoMessage {
		return &pb.OpenWrtConfig{}
	})
	runRoundTrip(t, "openvpn", "openvpn/*.json", func() protoMessage {
		return &pb.OpenVpnConfig{}
	})
	runRoundTrip(t, "wireguard", "wireguard/*.json", func() protoMessage {
		return &pb.WireguardConfig{}
	})
}

type protoMessage interface {
	proto.Message
	ValidateAll() error
}

func runRoundTrip(t *testing.T, name, glob string, factory func() protoMessage) {
	t.Helper()
	paths, err := filepath.Glob(glob)
	if err != nil {
		t.Fatalf("glob %s: %v", glob, err)
	}
	if len(paths) == 0 {
		t.Fatalf("no files matched %s", glob)
	}

	for _, p := range paths {
		t.Run(name+"/"+filepath.Base(p), func(t *testing.T) {
			raw, err := os.ReadFile(p)
			if err != nil {
				t.Fatalf("read file: %v", err)
			}
			want := normalizeJSON(t, raw)

			msg := factory()
			unmarshalOpts := protojson.UnmarshalOptions{DiscardUnknown: true}
			if err := unmarshalOpts.Unmarshal(raw, msg); err != nil {
				t.Fatalf("unmarshal proto: %v", err)
			}
			if err := msg.ValidateAll(); err != nil {
				t.Fatalf("validate proto: %v", err)
			}

			marshalOpts := protojson.MarshalOptions{UseProtoNames: true}
			gotRaw, err := marshalOpts.Marshal(msg)
			if err != nil {
				t.Fatalf("marshal proto: %v", err)
			}
			got := normalizeJSON(t, gotRaw)
			if !bytes.Equal(want, got) {
				t.Fatalf("round-trip mismatch:\nexpected: %s\nactual:   %s", want, got)
			}
		})
	}
}

func normalizeJSON(t *testing.T, data []byte) []byte {
	t.Helper()
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("normalize json: %v", err)
	}
	v = pruneEmpty(v)
	normalized, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal normalized json: %v", err)
	}
	return normalized
}

func pruneEmpty(v any) any {
	switch val := v.(type) {
	case map[string]any:
		for k, entry := range val {
			val[k] = pruneEmpty(entry)
			if arr, ok := val[k].([]any); ok && len(arr) == 0 {
				delete(val, k)
			}
		}
	case []any:
		for i, entry := range val {
			val[i] = pruneEmpty(entry)
		}
	}
	return v
}
