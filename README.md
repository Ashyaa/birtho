# Birtho
Halloween "Trick or Treat"-like bot in Go

# Done
## Commands
- Commands to configure the bot:
  - Choose who the admins are
  - Choose in which channels the bot will make items appear
  - Choose the command prefix
  - Toggle the game on or off
- Command that shows the current configuration of the bot
- Commands can be used with the configured prefix (eg `a!info`) or with a mention to the bot (eg `@bot info`)
- Command to reset the game
- Command to configure the minimum and maximum cooldown for monster spawns

## Basic game features
- The list of monsters and items the bot will use is read from a YAML configuration file, not provided in the repository (see YAML Configuration)
  - Each monster has a list of items that it can give to players
  - Each monster can be given a chance to spawn, else all monsters have the same chance to spawn
  - Each item a monster can give can have a chance to be given, else all items have the same chance
- Monsters appear when user post messages in the configured channels
- Users grab an item with a dedicated command, only the first user grabs the item
- If no one grabs the item within a few seconds, it disappears
- Whether or not it was grabbed by a user, a delay is put in place before another item appears
- The bot keeps in memory which items were grabbed by each user; repeats do not count
- The goal is to get all the items, the first player to do so is declared the winner

# TODO
## Scoreboard commands:
- By user
- For the whole server

## A `how to play` command
- By DM
- On the server itself

## Game
- Trick or Treat system: users have to use the correct command for each monster
  - Random or fixed by monster?
- Score system
- Winner announcement

## Configuration
- Make appearance duration configurable