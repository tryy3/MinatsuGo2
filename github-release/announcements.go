package github_release

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/cache/v9"
	"github.com/shurcooL/githubv4"
	realtimego "github.com/tryy3/realtime-go"
)

type AnnouncementReceiver struct {
	ChannelID      string
	ReleaseVersion string
}

type AnnouncementManager struct {
	// receivers    []*AnnouncementReceiver // Direct PMs
	discord      *Discord
	database     *Database
	githubClient *githubv4.Client
	cache        *cache.Cache
}

func (manager *AnnouncementManager) AddNewReceiver(recv *AnnouncementReceiver) error {
	key := fmt.Sprintf("receivers-%s", recv.ReleaseVersion)

	var existingReceivers []*AnnouncementReceiver
	err := manager.cache.Get(context.Background(), key, &existingReceivers)
	if err != nil && err != cache.ErrCacheMiss {
		return fmt.Errorf("error getting receivers from cache: %w", err)
	}

	if existingReceivers == nil {
		existingReceivers = []*AnnouncementReceiver{}
	}

	existingReceivers = append(existingReceivers, recv)

	err = manager.cache.Set(&cache.Item{
		Ctx:   context.Background(),
		Key:   key,
		Value: existingReceivers,
		TTL:   time.Hour,
	})
	if err != nil {
		return fmt.Errorf("error setting receivers in cache: %w", err)
	}
	return nil
}

func (manager *AnnouncementManager) Close() {
	manager.discord.Close()
}

func (manager *AnnouncementManager) getGithubVersions() []string {
	var latestVersion []string
	err := manager.cache.Get(context.Background(), "latestGithubVersion", &latestVersion)
	if err != nil && err != cache.ErrCacheMiss {
		log.Fatalf("error getting receivers from cache: %v", err)
	}

	if len(latestVersion) > 0 {
		return latestVersion
	}

	semverVersion, err := getNextGithubReleaseVersion(manager.githubClient)
	if err != nil {
		log.Fatalf("error getting next github version %v", err)
	}

	value := []string{semverVersion.String()}

	err = manager.cache.Set(&cache.Item{
		Ctx:   context.Background(),
		Key:   "latestGithubVersion",
		Value: value,
		TTL:   time.Hour,
	})

	if err != nil {
		log.Printf("error setting version in cache: %v", err)
		return []string{}
	}

	return value
}

func (manager *AnnouncementManager) githubNewReleaseCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// user ID
	var user *discordgo.User
	var err error
	if i.User != nil {
		user = i.User
	} else {
		user = i.Member.User
	}
	userID := user.ID
	fmt.Printf("UserID: %v\n", userID)
	fmt.Printf("User: %#v\n", i.User)
	fmt.Printf("Member: %#v\n", i.Member)

	// TODO: Check if this is a valid semver version
	param := i.ApplicationCommandData().Options[0].Options[0].Options[0].Value
	version := fmt.Sprintf("v%s", param)

	repoID, commitID, err := getGithubRepoAndCommit(manager.githubClient, "main")
	if err != nil {
		return fmt.Errorf("error getting github repo and commit: %w", err)
	}

	err = createNewGithubReleaseVersion(manager.githubClient, repoID, version, commitID)
	if err != nil {
		return fmt.Errorf("error creating new github release: %w", err)
	}

	err = manager.AddNewReceiver(&AnnouncementReceiver{ChannelID: userID, ReleaseVersion: version})
	if err != nil {
		return fmt.Errorf("error adding new receiver: %w", err)
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Tag %s has been created and a new release will soon be ready. I will let you know as soon as it is ready to edit.", version),
		},
	})
	if err != nil {
		return fmt.Errorf("error replying to command: %w", err)
	}

	return nil
}

func (manager *AnnouncementManager) githubNewReleaseCommandAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var choices []*discordgo.ApplicationCommandOptionChoice
	var versions = manager.getGithubVersions()
	for _, v := range versions {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  v,
			Value: v,
		})
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		log.Fatalf("error sending back option values: %v", err)
	}
}

func (manager *AnnouncementManager) GithubNewReleaseCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		err := manager.githubNewReleaseCommandHandler(s, i)
		if err != nil {
			log.Fatalf("error handling github new release command: %v", err)
		}
	case discordgo.InteractionApplicationCommandAutocomplete:
		manager.githubNewReleaseCommandAutocomplete(s, i)
	}
}

func (manager *AnnouncementManager) DatabaseUpdate(m realtimego.Message) {
	log.Printf("Topic: %s", m.Topic)
	if !strings.Contains(string(m.Topic), "events") {
		return
	}

	payload := m.Payload.(map[string]interface{})
	record := payload["record"].(map[string]interface{})
	eventName := record["event"].(string)
	log.Printf("Event: %s", eventName)
	if eventName != "release" {
		return
	}

	id := record["id"].(float64)

	event, err := manager.database.GetEvent(id)
	if err != nil {
		log.Fatalf("error getting event from database: %v", err)
	}

	tagName := event.RawData["release"].(map[string]interface{})["tag_name"].(string)
	url := event.RawData["release"].(map[string]interface{})["html_url"].(string)

	var newReceivers []*AnnouncementReceiver
	err = manager.cache.Get(context.Background(), fmt.Sprintf("receivers-%s", tagName), &newReceivers)
	if err != nil && err != cache.ErrCacheMiss {
		log.Fatalf("error getting receivers from cache: %v", err)
	}

	for _, recv := range newReceivers {
		log.Printf("Receiver version: %v - Tag name: %v", recv.ReleaseVersion, tagName)
		if recv.ReleaseVersion == tagName {
			log.Printf("Sending message to %v", recv.ChannelID)
			channel, err := manager.discord.session.UserChannelCreate(recv.ChannelID)
			if err != nil {
				log.Fatalf("error creating user channel: %v", err)
			}
			manager.discord.session.ChannelMessageSend(channel.ID, fmt.Sprintf("The release %s is now ready to be edited. You can find it at %s", tagName, url))
		}
	}

	if len(newReceivers) > 0 {
		err = manager.cache.Delete(context.Background(), fmt.Sprintf("receivers-%s", tagName))
		if err != nil {
			log.Fatalf("error deleting receivers from cache: %v", err)
		}
	}
}

func newGitHubAuth() (*http.Client, error) {
	// GitHub App installation
	appID, err := strconv.ParseInt(os.Getenv("APP_ID"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing APP_ID: %w", err)
	}

	installationID, err := strconv.ParseInt(os.Getenv("APP_INSTALLATION_ID"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing APP_INSTALLATION_ID: %w", err)
	}

	pemFile := os.Getenv("APP_PEM_FILE")
	if pemFile == "" {
		return nil, fmt.Errorf("APP_PEM_FILE is empty")
	}

	log.Printf("Pulling credentials for GitHub App\n")
	// itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, appID, installationID, pemFile)
	itr, err := ghinstallation.New(http.DefaultTransport, appID, installationID, []byte(pemFile))
	if err != nil {
		return nil, fmt.Errorf("error getting installation token: %w", err)
	}

	return &http.Client{Transport: itr}, nil
}

func NewAnnouncementManager() *AnnouncementManager {
	manager := &AnnouncementManager{}
	// manager.versionCache = cache.New(1*time.Minute, 10*time.Minute)
	cache, err := newRedisInstance()
	if err != nil {
		log.Fatalf("error initializing cache: %v", err)
	}
	manager.cache = cache

	discordSession, err := NewDiscordSession()
	if err != nil {
		log.Fatalf("error initializing discord: %v", err)
	}
	manager.discord = discordSession

	discordSession.AddNewCommand("", manager.GithubNewReleaseCommand, &discordgo.ApplicationCommand{
		Name:        "github",
		Description: "Subcommands and command groups example",
		Type:        discordgo.ChatApplicationCommand,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "new",
				Description: "Subcommands and command groups example",
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "release",
						Description: "Subcommands and command groups example",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:         discordgo.ApplicationCommandOptionString,
								Description:  "Subcommands and command groups example",
								Name:         "version",
								Autocomplete: true,
								Required:     true,
							},
						},
					},
				},
			},
		},
	})

	httpClient, err := newGitHubAuth()
	if err != nil {
		log.Fatalf("error getting github auth: %v", err)
	}

	manager.githubClient = githubv4.NewClient(httpClient)

	database, err := NewDatabase(manager)
	if err != nil {
		log.Fatalf("error initializing database: %v", err)
	}
	manager.database = database

	return manager
}
