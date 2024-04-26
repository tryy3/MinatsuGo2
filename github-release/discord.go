package github_release

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type Discord struct {
	session         *discordgo.Session
	commandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func (discord *Discord) AddNewCommand(guildID string, handler func(s *discordgo.Session, i *discordgo.InteractionCreate), cmd *discordgo.ApplicationCommand) {
	discord.commandHandlers[cmd.Name] = handler
	_, err := discord.session.ApplicationCommandCreate(discord.session.State.User.ID, "", cmd)
	if err != nil {
		log.Panicf("Cannot create '%v' command: %v", cmd.Name, err)
	}
}

func (discord *Discord) Close() {
	discord.session.Close()
}

func NewDiscordSession() (*Discord, error) {
	// phoneix realtime server endpoint
	discordToken := getSecretData("DISCORD_TOKEN")

	s, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
	discord := &Discord{session: s, commandHandlers: map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){}}
	discord.session.Identify.Intents = discordgo.IntentsGuildMessages

	discord.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		fmt.Printf("type: %#v\n", i.Type)
		fmt.Printf("name: %#v\n", i.ApplicationCommandData().Name)
		if h, ok := discord.commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	discord.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	err = discord.session.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	return discord, nil
}
