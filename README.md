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

## Basic game features
- Scalable list of items the bot will use through a JSON configuration file, not provided in the repository (see JSON Configuration)
- Items appear when user post messages in the configured channels
- Users grab an item with a dedicated command, only the first user grabds the item
- If no one grabs the item within a few seconds, it disappears
- Whether or not it was grabbed by a user, a delay is put in place before another item appears
- The bot keeps in memory which items were grabbed by each user; repeats do not count

# TODO
## Scoreboard commands:
- By user
- For the whole server

## A `how to play` command
- By DM
- On the server itself

## Game
- Monster system: have a list of monster with a image bound, than can provide a fixed list of items
- Spawn probability system
  - for monsters?
  - for items?
- Trick or Treat system: users have to use the correct command for each monster
  - Random or fixed by monster?
- Score system
- Winner announcement?

## Configuration
- Make cooldown configurable
- Make appearance duration configurable