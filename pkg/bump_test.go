package bump

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCastToVersion(t *testing.T) {
	tcs := []struct {
		input       string
		wantVersion Version
		wantErr     error
	}{
		{
			"v0.0.0",
			Version{majorRelease: 0, minorRelease: 0, bugFixRelease: 0},
			nil,
		},
		{
			"v143.73234.12",
			Version{majorRelease: 143, minorRelease: 73234, bugFixRelease: 12},
			nil,
		},
		{
			"0.0.0",
			Version{majorRelease: 0, minorRelease: 0, bugFixRelease: 0},
			ErrVersionFormat,
		},
		{
			"v.73234.12",
			Version{majorRelease: 0, minorRelease: 0, bugFixRelease: 0},
			ErrVersionFormat,
		},
		{
			"v..12",
			Version{majorRelease: 0, minorRelease: 0, bugFixRelease: 0},
			ErrVersionFormat,
		},
		{
			"1.73234.12",
			Version{majorRelease: 0, minorRelease: 0, bugFixRelease: 0},
			ErrVersionFormat,
		},
		{
			"73234.12",
			Version{majorRelease: 0, minorRelease: 0, bugFixRelease: 0},
			ErrVersionFormat,
		},
		{
			"12",
			Version{majorRelease: 0, minorRelease: 0, bugFixRelease: 0},
			ErrVersionFormat,
		},
		{
			"a.b.c",
			Version{majorRelease: 0, minorRelease: 0, bugFixRelease: 0},
			ErrVersionFormat,
		},
		{
			"",
			Version{majorRelease: 0, minorRelease: 0, bugFixRelease: 0},
			ErrVersionFormat,
		},
	}
	for _, c := range tcs {
		t.Run(fmt.Sprintf("input: %s, want version: %+v, want error: %v", c.input, c.wantVersion, c.wantErr), func(t *testing.T) {
			gotVersion, gotErr := CastToVersion(c.input)

			assert.ErrorIs(t, gotErr, c.wantErr)
			assert.Equal(t, c.wantVersion, gotVersion)
		})
	}
}
func TestVersion(t *testing.T) {
	t.Run("test String", func(t *testing.T) {
		v := &Version{majorRelease: 3, minorRelease: 5, bugFixRelease: 12}
		got := v.String()
		assert.Equal(t, "v3.5.12", got)
	})
	t.Run("test incrementMajor", func(t *testing.T) {
		v := &Version{majorRelease: 3, minorRelease: 5, bugFixRelease: 12}
		v.incrementMajor()
		assert.Equal(t, "v4.0.0", v.String())
	})
	t.Run("test incrementMinor", func(t *testing.T) {
		v := &Version{majorRelease: 3, minorRelease: 5, bugFixRelease: 12}
		v.incrementMinor()
		assert.Equal(t, "v3.6.0", v.String())
	})
	t.Run("test incrementBugFix", func(t *testing.T) {
		v := &Version{majorRelease: 3, minorRelease: 5, bugFixRelease: 12}
		v.incrementBugFix()
		assert.Equal(t, "v3.5.13", v.String())
	})
}
func TestValidateVersionFormat(t *testing.T) {
	tcs := []struct {
		input string
		want  error
	}{
		{"v0.0.0", nil},
		{"v143.73234.12", nil},
		{"0.0.0", ErrVersionFormat},
		{"v.73234.12", ErrVersionFormat},
		{"v..12", ErrVersionFormat},
		{"1.73234.12", ErrVersionFormat},
		{"73234.12", ErrVersionFormat},
		{"12", ErrVersionFormat},
		{"a.b.c", ErrVersionFormat},
		{"", ErrVersionFormat},
	}
	for _, c := range tcs {
		t.Run(fmt.Sprintf("input: %s, want: %v", c.input, c.want), func(t *testing.T) {
			got := validateVersionFormat(c.input)
			assert.Equal(t, c.want, got)
		})
	}
}
func TestGetLatestVersionTag(t *testing.T) {
	tcs := []struct {
		input       []string
		wantVersion Version
		wantErr     error
	}{
		{[]string{"v0.0.0"}, Version{0, 0, 0}, nil},
		{[]string{"v143.73234.12"}, Version{143, 73234, 12}, nil},
		{[]string{"0.0.0"}, Version{0, 0, 0}, ErrNoVersionTagsFound},
		{[]string{"v.73234.12"}, Version{0, 0, 0}, ErrNoVersionTagsFound},
		{[]string{"v..12"}, Version{0, 0, 0}, ErrNoVersionTagsFound},
		{[]string{"1.73234.12"}, Version{0, 0, 0}, ErrNoVersionTagsFound},
		{[]string{"73234.12"}, Version{0, 0, 0}, ErrNoVersionTagsFound},
		{[]string{"12"}, Version{0, 0, 0}, ErrNoVersionTagsFound},
		{[]string{"a.b.c"}, Version{0, 0, 0}, ErrNoVersionTagsFound},
		{[]string{""}, Version{0, 0, 0}, ErrNoVersionTagsFound},

		{[]string{"v0.0.0", "v0.0.1", "v0.1.0", "v1.0.0"}, Version{1, 0, 0}, nil},
		{[]string{"v0.0.25", "v0.0.5"}, Version{0, 0, 25}, nil},
		{[]string{"v0.0.25", "v1.0.5"}, Version{1, 0, 5}, nil},
	}
	for i, c := range tcs {
		t.Run(fmt.Sprintf("Test %d, input: %v, wantVersion: %v, wantErr: %v", i, c.input, c.wantVersion, c.wantErr), func(t *testing.T) {
			gotVersion, gotErr := getLatestVersionTag(c.input)

			assert.Equal(t, c.wantVersion, gotVersion)
			assert.ErrorIs(t, gotErr, c.wantErr)
		})
	}
}
func TestIncrementVersion(t *testing.T) {
	tcs := []struct {
		major       bool
		minor       bool
		wantVersion Version
		wantErr     error
	}{
		{
			major: false,
			minor: false,
			wantVersion: Version{
				majorRelease:  5,
				minorRelease:  5,
				bugFixRelease: 6,
			},
			wantErr: nil,
		},
		{
			major: true,
			minor: false,
			wantVersion: Version{
				majorRelease:  6,
				minorRelease:  0,
				bugFixRelease: 0,
			},
			wantErr: nil,
		},
		{
			major: false,
			minor: true,
			wantVersion: Version{
				majorRelease:  5,
				minorRelease:  6,
				bugFixRelease: 0,
			},
			wantErr: nil,
		},
		{
			major: true,
			minor: true,
			wantVersion: Version{
				majorRelease:  5,
				minorRelease:  5,
				bugFixRelease: 5,
			},
			wantErr: ErrCannotIncrementMajAndMin,
		},
	}
	for i, c := range tcs {
		t.Run(fmt.Sprintf("Test %d: major: %t, minor: %t, wantVersion: %v, wantErr: %v", i, c.major, c.minor, c.wantVersion, c.wantErr), func(t *testing.T) {
			v := Version{
				majorRelease:  5,
				minorRelease:  5,
				bugFixRelease: 5,
			}
			newVersion, err := incrementVersion(v, c.major, c.minor)

			assert.Equal(t, c.wantVersion, newVersion)
			assert.ErrorIs(t, err, c.wantErr)
		})
	}
}
