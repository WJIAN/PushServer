
local cid = KEYS[1]

local msgid = ARGV[1]
local stamp = tonumber(ARGV[2])

local rmk = 'RM.'..cid

if redis.call('ZSCORE', rmk, msgid) then
   -- 是重复的消息，直接返回0
   return 1
end


local num = redis.call('ZCARD', rmk)
-- < 20直接存储，不删除老的
if num >= 20 then
   -- >=20个的要检查所有超过20并多于600s的
   local rmcn = redis.call('ZCOUNT', rmk, 0, stamp-600)
   if rmcn > num-20 then rmcn = num-20 end
   if rmcn > 0 then
      redis.call('ZREMRANGEBYRANK', rmk, 0, rmcn-1)
   end
end

num = redis.call('ZCARD', rmk)
if num >= 100 then
   -- 超过了100，不管是不是10分的都干掉
   redis.call('ZREMRANGEBYRANK', rmk, 0, num-100)
end




redis.call('ZADD', rmk, stamp, msgid)


return 0

