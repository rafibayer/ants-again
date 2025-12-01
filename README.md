go tool pprof cpu.prof
go tool pprof -dot cpu.prof > cpu.dot

depending on params, seeing 1 of 2 familiar problems
- either, the ants don't care enough about the path, and they wander forever
- or, they care too much and follow each other in circles 

I think:
- wider search radius would help, along with stronger weighting for closer and angular similarity
- maybe back to kd tree for pheromone, but less much less density, e.g. only deposit every N frames
  - this combined with large radius might still be effective?
  - thinking about this sim: https://www.reddit.com/r/rust/comments/15dp0hq/media_ant_colony_simulation_in_rust_and_bevy/


things working well now, need a lot of perf work though
- some in rendering, some in management of kdtrees

big finds:
- larger search radius matters a lot
- weighting signals based on distance, cosine similarity


perf opportunities
- less frequent sensing, more influence

apparently drawing food as circles was incredibly expensive... how did I miss this with pprof?
anti-aliasing was also very expensive, not needed