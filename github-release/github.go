package github_release

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/shurcooL/githubv4"
)

const (
	github_owner      = "tryy3"
	github_repository = "flutter-game-jam-2023"
)

func getGithubVersionsFromGithub(client *githubv4.Client, count int32) ([]string, error) {
	var query struct {
		Repository struct {
			Description string
			Name        string
			Releases    struct {
				Nodes []struct {
					ID  string
					Tag struct {
						Name string
					}
				}
			} `graphql:"releases(first:$count, orderBy: $orderBy)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	err := client.Query(context.Background(), &query, map[string]interface{}{
		"owner":   githubv4.String(github_owner),
		"name":    githubv4.String(github_repository),
		"count":   githubv4.Int(count),
		"orderBy": githubv4.ReleaseOrder{Field: githubv4.ReleaseOrderFieldCreatedAt, Direction: githubv4.OrderDirectionDesc},
	})
	if err != nil {
		return nil, fmt.Errorf("error query github: %v", err)
	}

	versions := []string{}
	for _, node := range query.Repository.Releases.Nodes {
		versions = append(versions, node.Tag.Name)
	}
	return versions, nil
}

func getGithubRepoAndCommit(client *githubv4.Client, refName string) (string, string, error) {
	var query struct {
		Repository struct {
			ID   githubv4.ID
			Name string

			Ref struct {
				Target struct {
					Commit struct {
						OID githubv4.ID
					} `graphql:"... on Commit"`
				}
			} `graphql:"ref(qualifiedName: $refName)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	err := client.Query(context.Background(), &query, map[string]interface{}{
		"owner":   githubv4.String(github_owner),
		"name":    githubv4.String(github_repository),
		"refName": githubv4.String(refName),
	})
	if err != nil {
		return "", "", fmt.Errorf("error query github: %v", err)
	}

	return query.Repository.ID.(string), query.Repository.Ref.Target.Commit.OID.(string), nil
}

func getNextGithubReleaseVersion(client *githubv4.Client) (*semver.Version, error) {
	versions, err := getGithubVersionsFromGithub(client, 1)
	if err != nil {
		return nil, fmt.Errorf("error retrieving versions: %v", err)
	}

	version, err := semver.NewVersion(versions[0])
	if err != nil {
		return nil, fmt.Errorf("error parsing version: %v", err)
	}

	newVersion := version.IncMinor()
	return &newVersion, nil
}

func createNewGithubReleaseVersion(client *githubv4.Client, repositoryID string, name string, oid string) error {
	var mutation struct {
		CreateRef struct {
			Ref struct {
				ID     githubv4.ID
				Name   string
				Prefix string
			}
		} `graphql:"createRef(input: $input)"`
	}

	var ref = fmt.Sprintf("refs/tags/%s", name)

	input := githubv4.CreateRefInput{
		RepositoryID: githubv4.String(repositoryID),
		Name:         githubv4.String(ref),
		Oid:          githubv4.GitObjectID(oid),
	}

	err := client.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		return fmt.Errorf("error creating release: %v", err)
	}

	return nil
}
