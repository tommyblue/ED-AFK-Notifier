# Elite Dangerous AFK notifier

This tool can help Elite Dangerous Commanders when playing with their
[AFK Build](#what-is-afk-in-elite-dangerous).

The tool monitors the game [journal file](http://edcodex.info/?m=doc) for
hull damages and, in that case, immediately sends a
[Telegram](https://telegram.org/) message to alert the Commander.

- [Elite Dangerous AFK notifier](#elite-dangerous-afk-notifier)
  - [What is AFK in Elite Dangerous](#what-is-afk-in-elite-dangerous)
    - [How can you earn with AFK?](#how-can-you-earn-with-afk)
  - [Features](#features)
  - [Usage](#usage)
  - [Configuration](#configuration)
  - [How to create the Telegram bot](#how-to-create-the-telegram-bot)
  - [How to configure Gotify for Notifications](#how-to-configure-gotify-for-notifications)
    - [Prerequisites](#prerequisites)
    - [Configuration Steps](#configuration-steps)
    - [Example Configuration](#example-configuration)
    - [Testing the Configuration](#testing-the-configuration)
    - [Additional Notes](#additional-notes)
  - [All the details about AFK](#all-the-details-about-afk)
    - [Ship build](#ship-build)
    - [Stacking missions](#stacking-missions)
    - [Before you leave the station](#before-you-leave-the-station)
    - [Where to go](#where-to-go)
    - [Setup once parked](#setup-once-parked)
    - [What to expect once parked](#what-to-expect-once-parked)

## What is AFK in Elite Dangerous

AFK, a.k.a. "Away From Keyboard" is a very efficient way to gain credits in Elite Dangerous while
not actively playing.

The original idea comes from the [Hawkes Gaming](https://www.youtube.com/channel/UCwNphtRn-iP0HEZ98YLHAuA)
YouTube video ["How to AFK Your Way to Billions in 2021 Elite Dangerous Money Making Guide Solo Interview"](https://www.youtube.com/watch?v=aEv7K8ml3YY).

The idea is to build a highly engineered Type-10 ship, stack a lot of massacre missions, put some valuable
commodity in your cargo and park the ship in a low-resource extraction site.

Then leave the game there and do whatever you want (work, sleep, etc.).

You will be soon attacked by pirates that will be easily killed without suffering any damage.
After some hours all missions will be completed and you can go to get the rewards.

### How can you earn with AFK?

After some AFK rounds your reputation with parties will be very high and they'll give missions with rewards up to 40 million (with an average of 30M).

So, with 20 missions the total reward is around 600M + the bounties (30 to 60 million in my experience).

With a fully engineered T10 and a deadly/elite fighter, it takes around 4/6 hours to complete all 20 missions,
which means it's very easy to gain around **1.5 billion per day** (with 3 AFK rounds per day).

## Features

Send Telegram or Gotify messages on:

* Ship shields going down/up
* Ship hull damages
* Ship destroyed
* All missions are completed
* Fighter hull damage (optional)
* Total earned credits and pirates destroyed (optional)

## Usage

[Download the binary](https://github.com/tommyblue/ED-AFK-Notifier/releases) for your operating system
(Windows or Linux), place it somewhere and create, in the same folder, a file named `config.toml`
(see below for its content).

If you have problems extracting the release file (tar.gz), use [7-zip](https://www.7-zip.org/download.html).

Run the program and you'll start getting notifications.

## Configuration

The `config.toml` file, that must be placed along with the downloaded binary, must have the
following content (check the [config.example.toml](./config.example.toml) file for the latest
version of the config file):

```toml
[journal]
    Path = "<path to the journal directory>"
    debug = false # Print a logline for each new line in the journal file
    fighter = false # When true, send notification also for hull damage to the fighter
    shields = true # When true, send a notification when shields state changes (up/down)
    kills = true # When true, send notification on each new kill, including total reward earned (noisy!)
    silent_kills = true # When true, reduce noise for kill notification, sending a notification every 10 kills

# Notification service, choose either telegram or gotify
[notification]
    service = "telegram" # Options: telegram, gotify
    
[telegram]
    token = "<bot token>"
    channelId = <channel ID>
```


The journal path (that is the folder where ED saves journal files) under Windows is like:
`C:\\Users\\<Your User>\\Saved Games\\Frontier Developments\\Elite Dangerous`, just replace
`<Your User>` with your username.

If you run the game under Linux with Steam Proton, the path is something like:
`/home/<username>/.local/share/Steam/steamapps/compatdata/<numeric id>/pfx/drive_c/users/steamuser/Saved Games/Frontier Developments/Elite Dangerous/"`
(edit `<username>` and `<numeric id>` accordingly to your installation).

Create a Telegram bot (see below) and replace `<bot token>` with the token you get from BotFather.

At this point, the `channelId` is still unknown but it is required to receive messages
from the bot.

**Don't worry**, the bot itself can send you the id.

Run the bot and send it a message with the text `/channel`. You'll receive the channel id in the
response message.  
Copy the value and replace `<channel ID>` in the configuration file with that value, then restart
the bot.  

You can send a `/check` message to verify the configuration. You should receive a message
from this tool.

![](channel_id.png)

## How to create the Telegram bot

[Creating a Telegram bot](https://core.telegram.org/bots#3-how-do-i-create-a-bot) is a simple task
as you can easily do it by sending messages with the BotFather bot. The screenshot below shows the
simple steps required to create a bot:

![](./botfather.png)

## How to configure Gotify for Notifications

### Prerequisites

1. A running Gotify server. You can set it up by following the [Gotify documentation](https://gotify.net/docs/install).
2. An application token generated in Gotify for authentication.

### Configuration Steps

1. Open your `config.toml` file located in the same directory as the ED-AFK-Notifier binary.
2. Update the `[notification]` section to use Gotify as the notification service:
   ```toml
   [notification]
   service = "gotify"
   ```

3. Add the `[gotify]` section with the following fields:
   ```toml
   [gotify]
   url = "https://gotify.example.com" # Replace with your Gotify server URL
   token = "<app token>" # Replace with your Gotify application token
   title = "Elite Dangerous" # (Optional) Title prefix for notifications
   priority = 5 # (Optional) Priority of the notification (1-10, default is 5)
   ```

   - **`url`**: The base URL of your Gotify server.
   - **`token`**: The application token generated in Gotify.
   - **`title`**: (Optional) A custom title for notifications. Defaults to "Elite Dangerous" if not provided.
   - **`priority`**: (Optional) The priority level of notifications. Defaults to `5` if not specified.

4. Save the `config.toml` file.

### Example Configuration

Here is an example configuration for Gotify:

```toml
[notification]
service = "gotify"

[gotify]
url = "https://gotify.myserver.com"
token = "abc123xyz"
title = "ED Alerts"
priority = 8
```

### Testing the Configuration

1. Run the ED-AFK-Notifier application.
2. Check your Gotify server for incoming notifications.
3. If you encounter issues, ensure the `url` and `token` fields are correct and that your Gotify server is accessible.

### Additional Notes

- The `priority` field can be adjusted based on the importance of the notifications. Higher values indicate higher priority.
- If the `title` field is not set, the default title "Elite Dangerous" will be used.

For further assistance, refer to the [Gotify documentation](https://gotify.net/docs).


## All the details about AFK

### Ship build

First, you need a T10 with a very strong shield (8A Prismatic, engineered with Reinforce G5, Fast charge).

The hardpoints must all be pulse lasers with Long-range G5 and any experimental effect (optional).  
To sustain the rate of fire you also need a Weapon-focused G5 Power Distributor.

The utilities are 2 Point Defence turrets on the bottom side of the ship (to defend your cargo, without cargo you don't get attacked)
and 6 shield boosters Heavy-Duty G5, Super Capacitors. At this point, you also need an Overcharged Power Plant.

Finally, some Hull Reinforcement Packages, Guardian Shield Reinforcement Packages, a Fighter Hangar, and some cargo.

You don't need to have all of this to start AFKing, but with lower engineering, you'll probably need to take care of your ship a little more
(and using this app you'll be always notified of any possible problem).

[This is my full build on Coriolis](https://s.orbis.zone/ime_). It's able to resist on the field for many sessions without a scratch.

### Stacking missions

The key point of AFK is to find a good system, i.e. a system with some stations (and at least an L landing pad)
giving a lot of massacre missions in a near system which must have at least a resource extraction site.

An example system is [Gliese 868](https://inara.cz/starsystem/?search=Gliese+868) giving missions to [LP 932-12](https://inara.cz/starsystem/?search=LP+932-12).

Luckily CMDR VicTic built a [very handy tool](https://edtools.cc/pve) for the job ([announcement](https://www.reddit.com/r/EliteDangerous/comments/hpzmox/psa_a_tool_for_finding_good_sources_of_pve/)).

Massacre missions can be found under the "Combat" mission filter. They can be either single or wing missions.  
Missions from different factions will stack, which means that a pirate kill is valid for all the missions, thus reducing the number of kills required to complete all the missions.

At the beginning, the mission reward is around 5 million for 30-50 pirates. When you become an ally of the factions the mission will give you up to 40 million for 30-80 pirates.

### Before you leave the station

Once you got all the 20 missions you're ready to leave the station. Remember these 2 things before leaving:

- buy a valuable commodity, like gold or platinum. I suggest a cargo capacity of 8 elements at least
- set your crew member as active: if you use a smaller ship, without crew members, to move around the system bases to get rewards and new missions, then once back in the T10, the crew member will be inactive

### Where to go

The best place to park the ship is a "Resource Extraction Site (Low)" in the target system because the rank of the pirates is low.

You can also try a "Resource Extraction Site" which is a bit riskier but works better (more pirates, though with higher rank).

Note that when going to the site you could be interdicted by pirates. Those pirates are generally master or higher.
I suggest submitting to the interception and destroying the pirate, or it will reach you in the extraction site where you'll also have to face the first wave of pirates.

### Setup once parked

Once arrived at the extraction site stop the ship, deploy the fighter and the hardpoints.

Go to the **Ship panel > Functions** and set **"Turret weapon mode"** to **"Fire at will"**.

A **very important** thing to do is to set the PIPs to:

- 3 SYS
- 3 WEP
- 0 ENG

Depending on your build you could also have more power on the shields:

- 4 SYS
- 2 WEP
- 0 ENG
 
An optional thing to do is to set **Ship panel > Pilot Preferences > Report Crimes Against Me** to **OFF**.
This is a personal choice: with this option as ON you get help from System Police, so it's a safer option.
But sometimes they kill new pirates without leaving you the possibility to fire at them, so increasing the time required to complete the missions.  
With the option as OFF, you get more pirates but it's a bit riskier.  
A good balance is to leave that as ON and change the fighter policy from "Defend" to "Attack at Will" so that
it moves farther from the ship following the pirates (though the fighter itself gets destroyed more often.)

### What to expect once parked

When you first park at the extraction site some pirates immediately attack the ship. That's the most
dangerous moment because you get attacked by 3-5 pirates simultaneously and a ship with low engineering could
suffer some damages. So, stay on your ship for some minutes.

After the first round of attacks two things could happen:

- you start receiving scans and attacks regularly, like 1 every 1-3 minutes. That's perfect and you can go AFK
- you don't receive any attack. That's something happening sometime. Just log out and log in again (don't forget to deploy your weapons!). It should be fixed then and you'll start receiving the first round of massive attacks followed by regular attacks
