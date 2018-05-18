package protocol

import (
	"commandinfo"
	"fmt"
	"nwmessage"
	"nwmodel"
	"sort"
	"strings"
)

// Recv ...
func (d *Dispatcher) Recv(m nwmodel.ClientMessage) error {
	if m.Data == "" {
		return nil
	}

	return commandList.Exec(d, m)
}

type lobbyCommand struct {
	commandinfo.Info
	handler func(*nwmodel.Player, *Dispatcher, []interface{}) error
}

type LobbyCommandGroup map[string]lobbyCommand

func (cg *LobbyCommandGroup) Exec(d *Dispatcher, m nwmodel.ClientMessage) error {
	fullCmd := strings.Split(m.Data, " ")

	// handle help
	if fullCmd[0] == "help" {
		if len(fullCmd) == 1 {

			m.Sender.Outgoing <- nwmessage.PsNeutral(cg.AllHelp())

		} else {
			help, err := cg.Help(fullCmd[1:])

			if err != nil {
				m.Sender.Outgoing <- nwmessage.PsError(err)
			}
			m.Sender.Outgoing <- nwmessage.PsNeutral(help)

		}
		return nil

	}

	// if we find the command, try to execute
	if cmd, ok := commandList[fullCmd[0]]; ok {

		args, err := cmd.ValidateArgs(fullCmd[1:])
		if err != nil {
			// if we have trouble validating args
			m.Sender.Outgoing <- nwmessage.PsError(fmt.Errorf("%s\nusage: %s", err.Error(), cmd.Usage()))
		} else {
			// otherwise actually execute the command
			err = cmd.handler(m.Sender, d, args)
			if err != nil {
				m.Sender.Outgoing <- nwmessage.PsError(err)
			}
		}

		return nil
	}

	// if we don't find the command, pass an error back to caller in case caller wants to do something else
	return unknownCommand(fullCmd[0])
}

// Help composes a help string for the given command
func (cg LobbyCommandGroup) Help(args []string) (string, error) {
	if cmd, ok := cg[args[0]]; ok {
		return cmd.LongHelp(), nil
	}
	return "", unknownCommand(args[0])
}

// func (cg LobbyCommandGroup) longestKey() int {
// 	length := 0
// 	for key := range cg {
// 		if len(key) > length {
// 			length = len(key)
// 		}
// 	}
// 	return length
// }

// AllHelp composes help for all commands in the group
func (cg LobbyCommandGroup) AllHelp() string {
	cmds := make([]string, len(cg))
	var i int
	for key := range cg {
		cmds[i] = key
		i++
	}

	sort.Strings(cmds)
	// offset := cg.longestKey()
	helpStr := make([]string, len(cmds)+1)
	helpStr[0] = "Available commands:"

	for i, cmd := range cmds {
		helpStr[i+1] = cg[cmd].ShortHelp()
	}

	return strings.Join(helpStr, "\n")
}

func unknownCommand(cmd string) error {
	return fmt.Errorf("Unknown command, '%s'", cmd)
}

// type playerCmd func(*nwmodel.Player, *Dispatcher, []string) nwmessage.Message

// var lobbyCmdList = map[string]playerCmd{
// 	// TODO leaveGame should demand confirmation
// 	"leave": cmdLeaveGame,

// 	"t":    cmdTell,
// 	"tell": cmdTell,

// 	"ls": cmdListGames,
// 	// "list": cmdListGames,

// 	"join": cmdJoinGame,

// 	"name": cmdSetName,

// 	"new": cmdNewGame,

// 	"rm": cmdKillGame,

// 	"who": cmdWho,

// 	"chat": cmdChat,
// }

// var globalCmdList = map[string]bool{
// 	// TODO leaveGame should demand confirmation
// 	"leave": true,
// 	"t":     true,
// 	"tell":  true,
// 	// "name":  true,
// }

// func (d *Dispatcher) Recv(m nwmodel.ClientMessage) error {
// 	for {

// 		msg := strings.Split(m.Data, " ")

// 		// if players not in a game (i.e in lobby)
// 		if game == nil {
// 			// PLAYER IN LOBBY

// 			// are we in chatmode?
// 			if p.ChatMode && msg[0] != "chat" {

// 				chatMsg := strings.Join(msg, " ")
// 				for _, player := range d.GetPlayers() {
// 					player.Outgoing <- nwmessage.PsChat(p.GetName(), "global", chatMsg)
// 				}

// 			} else {
// 				// if its a valid lobby command execute
// 				if handlerFunc, ok := lobbyCmdList[msg[0]]; ok {
// 					res := handlerFunc(p, d, msg[1:])
// 					if res.Data != "" {
// 						p.Outgoing <- res
// 					}
// 				} else {
// 					p.Outgoing <- nwmessage.PsError(fmt.Errorf("Unknown lobby command, '%s'", msg[0]))
// 				}
// 			}
// 			// if play is not in the middle of something (which in the lobby they never should be) send prompt
// 			p.SendPrompt()
// 		} else {
// 			// PLAYER IN GAME

// 			// is the command entered a global command?
// 			if _, ok := globalCmdList[msg[0]]; ok {
// 				// get the handler function
// 				handlerFunc := lobbyCmdList[msg[0]]
// 				res := handlerFunc(p, d, msg[1:])
// 				if res.Data != "" {
// 					p.Outgoing <- res
// 				}
// 			} else {
// 				// it's not a global command, let the game handle it
// 				game.Recv(m)
// 			}
// 		}
// 	}
// }

// func cmdChat(p *nwmodel.Player, d *Dispatcher, args []string) nwmessage.Message {
// 	if len(args) > 0 {
// 		// broadcast args
// 		msg := strings.Join(args, " ")

// 		for _, player := range d.GetPlayers() {
// 			player.Outgoing <- nwmessage.PsChat(p.GetName(), "global", msg)
// 		}
// 		return nwmessage.Message{}
// 	}

// 	p.ChatMode = !p.ChatMode

// 	var flag string
// 	if p.ChatMode {
// 		flag = "ON"
// 	} else {
// 		flag = "OFF"
// 	}

// 	return nwmessage.PsNeutral(fmt.Sprintf("ChatMode set to %s", flag))
// }

// func cmdSetName(p *nwmodel.Player, d *Dispatcher, args []string) nwmessage.Message {
// 	if d.locations[p] != nil {
// 		return nwmessage.PsError(errors.New("Can only change name while in Lobby"))
// 	}

// 	if len(args) < 1 {
// 		return nwmessage.PsError(errors.New("Expected 1 argument, received none"))
// 	}

// 	// check for name collision
// 	for _, player := range d.players {
// 		if args[0] == player.GetName() {
// 			return nwmessage.PsError(fmt.Errorf("The name '%s' is already taken", args[0]))
// 		}
// 	}

// 	p.SetName(args[0])
// 	// p.Outgoing <- nwmessage.PsPrompt(p.GetName() + "@lobby>")
// 	return nwmessage.PsSuccess("Name set to '" + p.GetName() + "'")
// }

// func cmdNewGame(p *nwmodel.Player, d *Dispatcher, args []string) nwmessage.Message {

// 	if len(args) == 0 || args[0] == "" {
// 		return nwmessage.PsError(errors.New("Need a name for new game"))
// 	}

// 	if _, ok := d.locations[p]; ok {
// 		return nwmessage.PsError(fmt.Errorf("You can't create a game. You're already in a game"))
// 	}

// 	// create the game
// 	err := d.createGame(nwmodel.NewDefaultModel(args[0]))

// 	if err != nil {
// 		return nwmessage.PsError(err)
// 	}

// 	err = d.joinRoom(p, args[0])
// 	if err != nil {
// 		return nwmessage.PsError(err)
// 	}

// 	// p.SendPrompt()
// 	return nwmessage.PsSuccess(fmt.Sprintf("New game, '%s', created and joined", args[0]))
// }

// func cmdKillGame(p *nwmodel.Player, d *Dispatcher, args []string) nwmessage.Message {

// 	if len(args) == 0 || args[0] == "" {
// 		return nwmessage.PsError(errors.New("Need a name for game to remove"))
// 	}

// 	game, ok := d.games[args[0]]

// 	if !ok {
// 		return nwmessage.PsError(fmt.Errorf("No game, '%s', exists", args[0]))
// 	}

// 	// check to make sure game is empty
// 	if len(game.GetPlayers()) != 0 {
// 		return nwmessage.PsError(fmt.Errorf("The game, '%s', is not empty", args[0]))
// 	}

// 	// clean up game
// 	// TODO is this sufficient?
// 	delete(d.games, args[0])

// 	return nwmessage.PsSuccess(fmt.Sprintf("The game, '%s', has been removed", args[0]))
// }

// func cmdWho(p *nwmodel.Player, d *Dispatcher, args []string) nwmessage.Message {
// 	// var location Room
// 	// location, ok := d.games[d.locations[p.ID]]

// 	// if !ok {
// 	// 	location = d.Lobby
// 	// }
// 	location := d.locations[p]
// 	if location == nil {
// 		location = d
// 	}

// 	if len(location.GetPlayers()) == 0 {
// 		return nwmessage.PsNeutral("There are no players here")
// 	}

// 	var playerNames sort.StringSlice

// 	for _, p := range location.GetPlayers() {
// 		playerNames = append(playerNames, p.GetName())
// 	}

// 	playerNames.Sort()

// 	retMsg := "Players here:\n" + strings.Join(playerNames, ", ")
// 	return nwmessage.PsNeutral(retMsg)
// }

// func cmdJoinGame(p *nwmodel.Player, d *Dispatcher, args []string) nwmessage.Message {
// 	if len(args) < 1 {
// 		return nwmessage.PsError(errors.New("Expected 1 argument, received none"))
// 	}

// 	_, ok := d.games[args[0]]
// 	if !ok {
// 		return nwmessage.PsError(fmt.Errorf("No game named '%s' exists", args[0]))
// 	}

// 	err := d.joinRoom(p, args[0])
// 	if err != nil {
// 		return nwmessage.PsError(err)
// 	}

// 	// p.SendPrompt()
// 	return nwmessage.PsSuccess(fmt.Sprintf("Joined game, '%s'", args[0]))
// }

// func cmdLeaveGame(p *nwmodel.Player, d *Dispatcher, args []string) nwmessage.Message {

// 	err := d.leaveRoom(p)

// 	if err != nil {
// 		return nwmessage.PsError(err)
// 	}

// 	p.Outgoing <- nwmessage.PsSuccess(fmt.Sprintf("You have left the game"))
// 	p.SendPrompt()
// 	return nwmessage.Message{}
// }

// func cmdTell(p *nwmodel.Player, d *Dispatcher, args []string) nwmessage.Message {

// 	if len(args) < 2 {
// 		return nwmessage.PsError(errors.New("Need a recipient and a message"))
// 	}

// 	var recip *nwmodel.Player

// 	// check lobby for recipient:
// 	for _, player := range d.players {
// 		if player.GetName() == args[0] {
// 			recip = player
// 		}
// 	}

// 	// if not in lobby check all games
// 	if recip == nil {
// 		for _, game := range d.games {
// 			for _, player := range game.GetPlayers() {
// 				if player.GetName() == args[0] {
// 					recip = player
// 				}
// 			}
// 		}
// 	}

// 	if recip == nil {
// 		return nwmessage.PsError(fmt.Errorf("No such player, '%s'", args[0]))
// 	}

// 	chatMsg := strings.Join(args[1:], " ")

// 	recip.Outgoing <- nwmessage.PsChat(p.GetName(), "private", chatMsg)
// 	return nwmessage.PsNeutral(fmt.Sprintf("(you to %s): %s", recip.GetName(), chatMsg))
// }

// func cmdListGames(p *nwmodel.Player, d *Dispatcher, args []string) nwmessage.Message {
// 	gameList := ""

// 	if len(d.games) == 0 {
// 		return nwmessage.PsNeutral("No games running. Type, 'new game_name', to start one")
// 	}

// 	for gameName, game := range d.games {
// 		gameList += fmt.Sprintf("'%s' - Players: %d\n", gameName, len(game.GetPlayers()))
// 	}

// 	return nwmessage.PsNeutral(strings.TrimSpace("Available games:\n" + gameList))
// }
