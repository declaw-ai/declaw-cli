package sandbox

import "testing"

func TestParseVolumeAttachments(t *testing.T) {
	t.Run("id and mount path only", func(t *testing.T) {
		got, err := parseVolumeAttachments([]string{"vol-1:/data"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("expected 1 attachment, got %d", len(got))
		}
		if got[0].VolumeID != "vol-1" || got[0].MountPath != "/data" {
			t.Errorf("unexpected attachment: %+v", got[0])
		}
		if got[0].Mode != "" || got[0].Subpath != "" {
			t.Errorf("expected empty mode/subpath, got %+v", got[0])
		}
	})

	t.Run("with mode", func(t *testing.T) {
		got, err := parseVolumeAttachments([]string{"vol-1:/data:mount-ro"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got[0].Mode != "mount-ro" {
			t.Errorf("expected mode=mount-ro, got %q", got[0].Mode)
		}
		if got[0].Subpath != "" {
			t.Errorf("expected empty subpath, got %q", got[0].Subpath)
		}
	})

	t.Run("with mode and subpath", func(t *testing.T) {
		got, err := parseVolumeAttachments([]string{"vol-1:/data:mount:sub/dir"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got[0].Mode != "mount" || got[0].Subpath != "sub/dir" {
			t.Errorf("unexpected attachment: %+v", got[0])
		}
	})

	t.Run("invalid mode rejected", func(t *testing.T) {
		if _, err := parseVolumeAttachments([]string{"vol-1:/data:bogus"}); err == nil {
			t.Fatal("expected error for invalid mode")
		}
	})

	t.Run("missing mount path rejected", func(t *testing.T) {
		if _, err := parseVolumeAttachments([]string{"vol-1"}); err == nil {
			t.Fatal("expected error for missing mount path")
		}
	})
}
