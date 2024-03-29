
local cid = KEYS[1]
local msgid = ARGV[1]
local pb = ARGV[2]
local stamp = tonumber(ARGV[3])

local ck = 'I.'..cid

local restaddr = redis.call('HGET', ck, 'restaddr')

if not restaddr then
   return "NOTFOUND"
end


local mk = 'M.'..string.sub(cid, 1, 5)..'.'..msgid
--local mk = 'M.'..cid..'.'..msgid
local smk = 'SM.'..cid

redis.call('SET', mk, pb)

-- 超过一个小时消息不发了
redis.call('EXPIRE', mk, 3600)

local num = redis.call('ZCARD', smk)

if num > 19 then
   -- 最多发20条没有超时的消息
   redis.call('ZREMRANGEBYRANK', smk, 0, num-20)

end

redis.call('ZADD', smk,
           stamp,
           msgid
)

-- 3600*24*7 = 604800
redis.call('EXPIRE', smk, 604800)

local timeout = redis.call('HGET', ck, 'timeout')
if not timeout then
   return "CLOSED"
end

if tonumber(timeout) == 0 then
   -- 说明正常完成了close动作
   local closestamp = redis.call('HGET', ck, 'closestamp')
   if not closestamp then
      return "CLOSED"
   end

   if stamp - tonumber(closestamp) > 300 then
      return "CLOSED"
   else
      -- 对于连接关闭小于5分的
      return "TMPCLOSED"
   end

end

if stamp > tonumber(timeout) then
   return "CLOSED"
else
   return restaddr
end
