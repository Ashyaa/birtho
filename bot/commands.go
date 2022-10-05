package bot

import (
	"fmt"
	"strconv"

	"github.com/ashyaa/birtho/util"
	DG "github.com/bwmarrin/discordgo"
)

func SendText(s *DG.Session, i *DG.Interaction, channelID, content string) (*DG.Message, error) {
	if i == nil {
		return s.ChannelMessageSend(channelID, content)
	}
	err := s.InteractionRespond(i, &DG.InteractionResponse{
		Type: DG.InteractionResponseChannelMessageWithSource,
		Data: &DG.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		return nil, err
	}
	return s.InteractionResponse(i)
}

func SendEmbed(s *DG.Session, i *DG.Interaction, channelID string, embed *DG.MessageEmbed) (*DG.Message, error) {
	if i == nil {
		return s.ChannelMessageSendEmbed(channelID, embed)
	}
	err := s.InteractionRespond(i, &DG.InteractionResponse{
		Type: DG.InteractionResponseChannelMessageWithSource,
		Data: &DG.InteractionResponseData{
			Embeds: []*DG.MessageEmbed{embed},
		},
	})
	if err != nil {
		return nil, err
	}
	return s.InteractionResponse(i)
}

func buildOptions(b *Bot) {
	for i := range b.Commands {
		if b.Commands[i].appCmd != nil {
			if b.Commands[i].Admin || b.Commands[i].AlwaysTrigger {
				b.Commands[i].appCmd.DefaultMemberPermissions = &DefaultMemberPermissions
			}
			b.Commands[i].appCmd.Options = b.Commands[i].DGOptions(b)
		}
	}
}

func (b *Bot) buildInteractionHandlers() {
	for _, cmd := range b.Commands {
		if cmd.appCmd == nil {
			continue
		}
		b.InteractionHandlers[cmd.Name] = HandlerFromInteraction(b, cmd)
	}
}

type CommandParameters struct {
	MID, UID, CID, GID, Name string // MID only available on MessageCreate, else empty
	I                        *DG.Interaction
	Options                  map[string]interface{}
	S                        Server
	IsUserTriggered          bool
}

func (p *CommandParameters) ParseOptionsFromRaws(raws []string, opts Options) error {
	if len(raws) < len(opts) {
		return fmt.Errorf("not enough arguments for command %s", p.Name)
	}
	for i, opt := range opts {
		raw := raws[i]
		switch opt.Type {
		case TypeString:
			p.Options[opt.Name] = raw
		case TypeInteger:
			v, err := strconv.Atoi(raw)
			if err != nil {
				return err
			}
			p.Options[opt.Name] = v
		case TypeChannel:
			v, ok := util.StripChannelTag(raw)
			if !ok {
				return fmt.Errorf("invalid channel: %s", raw)
			}
			p.Options[opt.Name] = v
		case TypeUser:
			v, ok := util.StripChannelTag(raw)
			if !ok {
				return fmt.Errorf("invalid channel: %s", raw)
			}
			p.Options[opt.Name] = v
		default:
			return fmt.Errorf("unknown option type: %s", opt.Type)
		}
	}
	return nil
}

func (p *CommandParameters) ParseOptionsFromInteraction(i *DG.Interaction, opts Options) error {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*DG.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	if len(options) < len(opts) {
		return fmt.Errorf("not enough arguments for command %s", p.Name)
	}
	for _, opt := range opts {
		dgOption, ok := optionMap[opt.Name]
		if !ok {
			return fmt.Errorf("missing option %s", opt.Name)
		}
		switch opt.Type {
		case TypeString:
			p.Options[opt.Name] = dgOption.StringValue()
		case TypeInteger:
			p.Options[opt.Name] = int(dgOption.IntValue())
		case TypeChannel:
			p.Options[opt.Name] = dgOption.ChannelValue(nil).ID
		case TypeUser:
			p.Options[opt.Name] = dgOption.UserValue(nil).ID
		default:
			return fmt.Errorf("unknown option type: %s", opt.Type)
		}
	}
	return nil
}

func ParamsFromInteraction(b *Bot, i *DG.InteractionCreate, name string) CommandParameters {
	serv := b.GetServer(i.GuildID)
	return CommandParameters{
		UID:             i.Member.User.ID,
		CID:             i.ChannelID,
		GID:             i.GuildID,
		I:               i.Interaction,
		Name:            name,
		Options:         map[string]interface{}{},
		S:               serv,
		IsUserTriggered: true,
	}
}

func ParamsFromMessageCreate(b *Bot, m *DG.MessageCreate, name string) CommandParameters {
	serv := b.GetServer(m.GuildID)
	return CommandParameters{
		MID:     m.Message.ID,
		UID:     m.Message.Author.ID,
		CID:     m.ChannelID,
		GID:     m.GuildID,
		I:       nil,
		Name:    name,
		Options: map[string]interface{}{},
		S:       serv,
	}
}

func HandlerFromMessageCreate(b *Bot, cmd Command) func(*DG.Session, *DG.MessageCreate) {
	return func(s *DG.Session, m *DG.MessageCreate) {
		if cmd.ModifiesServer {
			b.mutex.Lock()
			defer b.mutex.Unlock()
		}
		p := ParamsFromMessageCreate(b, m, cmd.Name)
		if p.UID == b.UserID {
			return
		}

		if cmd.Admin && !p.S.IsAdmin(p.UID) {
			return
		}

		raws, ok := b.triggered(s, m, p.S.Prefix, cmd.Name)
		if !ok && !cmd.AlwaysTrigger {
			return
		}
		p.IsUserTriggered = ok
		err := p.ParseOptionsFromRaws(raws, cmd.Options)
		if err != nil {
			SendText(s, nil, p.CID, err.Error())
			return
		}
		if !cmd.AlwaysTrigger {
			b.Info("command %s triggered", cmd.Name)
		}
		cmd.Action(b, p)
	}
}

func HandlerFromInteraction(b *Bot, cmd Command) func(*DG.Session, *DG.InteractionCreate) {
	return func(s *DG.Session, i *DG.InteractionCreate) {
		if cmd.ModifiesServer {
			b.mutex.Lock()
			defer b.mutex.Unlock()
		}
		p := ParamsFromInteraction(b, i, cmd.Name)

		serv := b.GetServer(p.GID)
		if cmd.Admin && !serv.IsAdmin(p.UID) {
			SendText(s, i.Interaction, p.CID, "Command not authorized")
			return
		}

		err := p.ParseOptionsFromInteraction(i.Interaction, cmd.Options)
		if err != nil {
			SendText(s, nil, p.CID, err.Error())
			return
		}
		if !cmd.AlwaysTrigger {
			b.Info("command %s triggered", cmd.Name)
		}
		cmd.Action(b, p)
	}
}

type OptionType string

const (
	TypeString  OptionType = "string"
	TypeInteger OptionType = "integer"
	TypeChannel OptionType = "channel"
	TypeUser    OptionType = "user"
)

type Option struct {
	Name, Description string
	Type              OptionType
}

type Options []Option

func dgOption(opt Option) (*DG.ApplicationCommandOption, error) {
	var typ DG.ApplicationCommandOptionType
	switch opt.Type {
	case TypeString:
		typ = DG.ApplicationCommandOptionString
	case TypeInteger:
		typ = DG.ApplicationCommandOptionInteger
	case TypeChannel:
		typ = DG.ApplicationCommandOptionChannel
	case TypeUser:
		typ = DG.ApplicationCommandOptionUser
	default:
		return nil, fmt.Errorf("unknown option type %s", opt.Type)
	}
	return &DG.ApplicationCommandOption{
		Name:        opt.Name,
		Description: opt.Description,
		Type:        typ,
		Required:    true,
	}, nil
}

func (c Command) DGOptions(b *Bot) []*DG.ApplicationCommandOption {
	res := []*DG.ApplicationCommandOption{}
	for _, opt := range c.Options {
		o, err := dgOption(opt)
		if err != nil {
			b.ErrorE(err, "building command option %s of command %s", opt.Name, c.Name)
			continue
		}
		res = append(res, o)
	}
	return res
}
