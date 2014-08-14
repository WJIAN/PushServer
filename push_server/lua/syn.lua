

local cid = KEYS[1]
local ck = 'I.'..cid
local smk = 'SM.'..cid

redis.call('HMSET', ck,
                  'remote', ARGV[1],
                  'appid', ARGV[2],
                  'installid', ARGV[3],
                  'restaddr', ARGV[4]
)

-- 3600*24*7 = 604800
redis.call('EXPIRE', ck, 604800)


--local msgsids = redis.call('ZREVRANGE', smk, -10, -1)
local msgsids = redis.call('ZRANGE', smk, 0, -1)

local msgs = {}
for i, v in ipairs(msgsids) do
   --local mk = 'M.'..cid..'.'..v
   local mk = 'M.'..string.sub(cid, 1, 5)..'.'..v
   if redis.call("TTL", mk) > 0 then
      msgs[#msgs+1] = v
      msgs[#msgs+1] = redis.call("GET", mk)
   else
      -- 超时的不要
      redis.call('ZREM', smk, v)
   end

end


return msgs


