-- export LUA_PATH='/home/pfsmagalhaes/.luarocks/share/lua/5.2/?.lua;/home/pfsmagalhaes/.luarocks/share/lua/5.2/?/init.lua;/usr/local/share/lua/5.2/?.lua;/usr/local/share/lua/5.2/?/init.lua;/usr/local/lib/lua/5.2/?.lua;/usr/local/lib/lua/5.2/?/init.lua;/usr/share/lua/5.2/?.lua;/usr/share/lua/5.2/?/init.lua;./?.lua'
-- export LUA_CPATH='/home/pfsmagalhaes/.luarocks/lib/lua/5.2/?.so;/usr/local/lib/lua/5.2/?.so;/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;/usr/lib/lua/5.2/?.so;/usr/local/lib/lua/5.2/loadall.so;./?.so'


local bconsumer = require "resty.kafka.basic-consumer"

local broker_list = { {host= "127.0.0.1",port= 9092,}}

client_config = nil

local c = bconsumer:new(broker_list, client_config)

result, err = c:fetch("teste_ordem", 0, 0)

if err ~= nil then
    print(err)
else
    print(result)
end
