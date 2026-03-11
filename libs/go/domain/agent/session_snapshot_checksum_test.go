package agent

import "testing"

func TestComputeSessionSnapshotChecksum_NormalizesEquivalentJSON(t *testing.T) {
	t.Parallel()

	first, err := ComputeSessionSnapshotChecksum(
		[]byte("{\"status\":\"running\",\"report\":{\"summary\":\"ok\"}}"),
		[]byte("{\n  \"session_id\": \"abc\",\n  \"step\": 1\n}"),
	)
	if err != nil {
		t.Fatalf("ComputeSessionSnapshotChecksum() error = %v", err)
	}

	second, err := ComputeSessionSnapshotChecksum(
		[]byte("{\"report\":{\"summary\":\"ok\"},\"status\":\"running\"}"),
		[]byte("{\"step\":1,\"session_id\":\"abc\"}"),
	)
	if err != nil {
		t.Fatalf("ComputeSessionSnapshotChecksum() error = %v", err)
	}

	if first != second {
		t.Fatalf("expected stable checksum, got %q and %q", first, second)
	}
}

func TestComputeSessionSnapshotChecksum_DistinguishesSnapshotPayloads(t *testing.T) {
	t.Parallel()

	first, err := ComputeSessionSnapshotChecksum(
		[]byte("{\"status\":\"running\"}"),
		[]byte("{\"session_id\":\"abc\"}"),
	)
	if err != nil {
		t.Fatalf("ComputeSessionSnapshotChecksum() error = %v", err)
	}

	second, err := ComputeSessionSnapshotChecksum(
		[]byte("{\"status\":\"succeeded\"}"),
		[]byte("{\"session_id\":\"abc\"}"),
	)
	if err != nil {
		t.Fatalf("ComputeSessionSnapshotChecksum() error = %v", err)
	}

	if first == second {
		t.Fatal("expected different checksums for different payloads")
	}
}
