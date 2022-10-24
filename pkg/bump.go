package bump

import (
	"fmt"
	"github.com/facette/natsort"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/pkg/errors"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

// Suggested that creating custom Error type is unnecessarily confusing, but without it I cannot make these errors constant.
const (
	ErrVersionFormat            = Error("error: version string formatted incorrectly")
	ErrNoTagsFound              = Error("error: no tags found on remote")
	ErrNoVersionTagsFound       = Error("error: no tags found have expected version format")
	ErrCannotIncrementMajAndMin = Error("error: pass either -minor OR -major flags, not both")
)

type (
	majorRelease  int
	minorRelease  int
	bugFixRelease int
)

func CastToVersion(version string) (Version, error) {
	if err := validateVersionFormat(version); err != nil {
		return Version{}, err
	}
	version = strings.TrimPrefix(version, "v")

	stringElements := strings.Split(version, ".")
	var intElements []int
	for _, s := range stringElements {
		i, err := strconv.Atoi(s)
		if err != nil {
			return Version{}, err
		}
		intElements = append(intElements, i)
	}

	return Version{majorRelease: majorRelease(intElements[0]), minorRelease: minorRelease(intElements[1]), bugFixRelease: bugFixRelease(intElements[2])}, nil
}

type Version struct {
	majorRelease  majorRelease
	minorRelease  minorRelease
	bugFixRelease bugFixRelease
}

func (v *Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.majorRelease, v.minorRelease, v.bugFixRelease)
}

func (v *Version) incrementMajor() {
	v.majorRelease++
	v.minorRelease = 0
	v.bugFixRelease = 0
}

func (v *Version) incrementMinor() {
	v.minorRelease++
	v.bugFixRelease = 0
}

func (v *Version) incrementBugFix() {
	v.bugFixRelease++
}

func BumpVersion(repo *git.Repository, major, minor bool) error {
	tagList, err := getRemoteGitTags(repo)
	if err != nil {
		return errors.Wrap(err, "Unable to get remote git tags from the repo")
	}
	latestVersionTag, err := getLatestVersionTag(tagList)
	if err != nil {
		if err == ErrNoVersionTagsFound {
			fmt.Println("no version tags found, creating new tag version: v0.0.0")
			if err := tag(repo, Version{}); err != nil {
				return err
			}
			return nil
		}
		return errors.Wrap(err, "unable to get latest version tag from the tag list")
	}
	fmt.Println("latest version tag at remote: ", latestVersionTag.String())

	incrementedVersionTag, err := incrementVersion(latestVersionTag, major, minor)
	if err != nil {
		return err
	}

	fmt.Printf("version tag incremented from %s to %s\n", latestVersionTag.String(), incrementedVersionTag.String())

	if err := tag(repo, incrementedVersionTag); err != nil {
		return err
	}

	return nil
}

// incrementVersion increments bug fix release by default. If minor or major flags are passed it will instead increment either of them.
func incrementVersion(latestVersion Version, major, minor bool) (Version, error) {
	incrementedVersion := latestVersion

	if major && minor {
		return latestVersion, ErrCannotIncrementMajAndMin
	}
	if major {
		incrementedVersion.incrementMajor()
		return incrementedVersion, nil
	}
	if minor {
		incrementedVersion.incrementMinor()
		return incrementedVersion, nil
	}
	incrementedVersion.incrementBugFix()
	return incrementedVersion, nil
}

// getRemoteGitTags retrieves all remote 'refs' from the git repo, then extracts the short name of each.
func getRemoteGitTags(repo *git.Repository) ([]string, error) {
	remote, err := repo.Remote("origin")
	if err != nil {
		return nil, errors.Wrap(err, "unable to access remote repo")
	}
	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list references from remote repo")
	}

	if len(refs) == 0 {
		return nil, ErrNoTagsFound
	}

	var tagList []string
	for _, r := range refs {
		tag := r.Name().Short()
		tagList = append(tagList, tag)
	}
	return tagList, nil

}

// getLatestVersionTag retrieves the latest version tag from the remote git tags.
func getLatestVersionTag(tagList []string) (Version, error) {
	var versionList []string
	for _, t := range tagList {
		if err := validateVersionFormat(t); err == nil {
			versionList = append(versionList, t)
		}
	}

	if len(versionList) == 0 {
		return Version{}, ErrNoVersionTagsFound
	}

	natsort.Sort(versionList)
	latestVersionStr := versionList[len(versionList)-1]
	latestVersion, err := CastToVersion(latestVersionStr)
	if err != nil {
		return Version{}, err
	}

	return latestVersion, nil
}

// validateVersionFormat ensures version adheres to the format "v0.0.0"
func validateVersionFormat(version string) error {
	if !regexp.MustCompile("^v[0-9]+\\.[0-9]+\\.[0-9]+$").MatchString(version) {
		return ErrVersionFormat
	}
	return nil
}

// tag creates a new tag and pushes all local tags to the remote repository.
func tag(repo *git.Repository, newTagVersion Version) error {
	fmt.Println("new tag: ", newTagVersion.String())

	h, err := repo.Head()
	if err != nil {
		return errors.Wrap(err, "unable to get HEAD from the opened repo")
	}

	if _, err := repo.CreateTag(newTagVersion.String(), h.Hash(), nil); err != nil {
		return errors.Wrap(err, "unable to create tag from the repo head hash and newly created tag version number. tag may already exist locally.")
	}
	fmt.Printf("new tag has been created locally (view with 'git tag -l | tail'): %s\n", newTagVersion.String())

	if err := repo.Push(&git.PushOptions{RemoteName: "origin", RefSpecs: []config.RefSpec{"refs/tags/*:refs/tags/*"}, Progress: os.Stderr}); err != nil {
		if rollbackErr := repo.DeleteTag(newTagVersion.String()); rollbackErr != nil {
			return errors.Wrap(rollbackErr, "push local tag to remote failed, so attempted to rollback local tag. rollback local tag failed. May require manual cleanup.")
		}
		return errors.Wrap(err, "unable to push new tag to remote repo")
	}
	return nil
}
