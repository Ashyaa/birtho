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
- Command to configure how long a monster stays before leaving

## Basic game features
- The list of monsters and items the bot will use is read from a YAML configuration file, not provided in the repository (see YAML Configuration)
  - Each monster has a list of items that it can give to players
  - Each monster can be given a chance to spawn, else all monsters have the same chance to spawn
  - Each item a monster can give can have a chance to be given, else all items have the same chance
- Monsters appear when user post messages in the configured channels
- Monsters drop an item when a user uses either the "trick" or the "treat" command. If the correct command is used, the user gets an item, else it maakes the monster leave. Whatever the result, only the first command is
aacknowledged, it's a matter of who is the fastest to type the command.
- If no one grabs the item within a few seconds, it disappears
- Whether or not it was grabbed by a user, a delay is put in place before another item appears
- The bot keeps in memory which items were grabbed by each user; repeats do not count
- The goal is to get all the items, the first player to do so is declared the winner

# TODO
## Scoreboard commands:
- By user: list of missing items
- For the whole server: score ranking for all players that have played (that have used trick or treat at least once)

## A `how to play` command
- By DM
- On the server itself

## Game
- 15 monsters; 1pt for a common item, 5 for uncommon, 10 for rare (240 points total)
- items spawns: 50% - 35% - 15%
- Winner announcement (give a role to the winner?)