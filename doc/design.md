first idea:
define some windows for the skill, latency gap and a soft / hard wait time limit
after the soft limit hits, the skill / latency window will increase in exp steps, so that when the hard limit happens, the player will be matched with p much anyone.

post-googling:<br>
[r/learnprogramming](https://www.reddit.com/r/learnprogramming/comments/7rdlzf/how_is_online_game_matchmaking_done_from_a/) - recommends doing it in n^2, as the queue size is usually not that large

after implementing the model.QueuedUser struct, i noticed that the user can be represented as a 3d point, with skill, latency and queuedAt as the axis. this turns the problem into a vertex grouping one.

here are some ideas:<br>
[cs.stackoverflow](https://cs.stackexchange.com/questions/85929/efficient-point-grouping-algorithm)
* spatial partitioning: would be fun, but building the graph every tick would be inefficient. storing the graph would be annyoing, because id need a special insert-delete optimized algorithm. basically, partitioning is a no go
* dbscan: never heard of it, but it seems to be implemented in go as a lib and the api seems easy-to-use so ill try this option, and see how it compares to the n^2 algo.

a problem of point representation: the axis have different scales and have to be normalized
ie: elo is between 1-1000, but the latency is 1-10000. in this case, elo will have 10x > importance over latency.
therefore, it would be highly preferrable to normalize the axis in some way, but: elo is technically infinite, and latency is also technically infinite (although practically caps at around 1s)
possible approaches:
a) normalize as InverseLerp(v, minV, maxV)
* easy to implement
* inconsistent results (dataset1 has latency 50-200, dataset2 has latency 50-1000, both have the same elo: latency ends up having different weights)
b) ceil the values (mb w/ an env var)
* very easy to implement
* consistent enough (all axis will have perfect weights, but there can be islands of bunched up points on the ceil edge)
c) ceil the values using smth like 1/x, so that theres still SOME difference between ie 1000 and 2000, even if the relationship is no longer linear
* annoying to implement (i am mostly worried about choosing a correct transformation function)
* mostly consistent results (values near the function ceil will become over-weighted, which might or might not be a problem.)

also, if normalization is done during the insert stage, it will be annoying to support the db during a ceil transfer, as every entry will have to be updated
therefore, normalization should definetely happen DURING matching and not before/after

my idea for queuedAt: increase its weight in the distance function as time passes

ok, idk why i didnt go with the grid space partitioning / bin algo:
its much simpler
its much faster AND memory efficient
it can be stored in the db (although, a db purge will be required if some grid parameters change)
and, an important note, it is easier to reason about in the temporal space: add stuff to bins, periodically check bins, if a user is waiting too long look into nearby bins, 