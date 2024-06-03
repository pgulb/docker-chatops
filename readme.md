# docker-chatops  
  
Telegram bot to execute docker commands using chat  
  
## configuration
configure it with `.env` file:  
`TELEGRAM_BOT_TOKEN`=your bot token from the BotFather  
`ALLOWED_CHAT_IDS`=comma-separated list of chats from which commands will work  
if you do not know your chat ID, run it without writing any, write something 
to bot and see your chat ID in logs    
see `example_env` for file template  
  
## running
docker socket must be exposed to container to allow it to interact with docker engine  
run it with:  
```sh
docker run -d --name chatops \
-v /var/run/docker.sock:/var/run/docker.sock \
-v ./.env:/app/.env:ro \
--restart unless-stopped ghcr.io/pgulb/docker-chatops:latest
```

## commands
It is best to create a list of those commands using BotFather
to use them more easily  
  
`/ps` - list all containers and their info  
`/logs` - tail logs of a container (30 lines max)  
`/restart` - restart specific container  
`/images` - show all images and metadata  
`/version` - display bot and docker engine version  
