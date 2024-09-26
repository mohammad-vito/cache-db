Note: This is a sample project intended to demonstrate my coding style.

The project has been developed with several assumptions, which is not reflective of how a real-world project should be approached. In a real scenario, the requirements should be clearly defined before development begins.

Potential Bug: Setting a key for an existing data entry may cause an issue. However, since the requirements were unclear regarding this case, I did not address it as a priority.

Any optimization should be based on profiling and identifying bottlenecks. I have deliberately avoided premature optimization and over-engineering in this project. but here is an example of how we can imrpove the performance of this system

Improvements:
- If the system encounters excessive lock contention, we can shard the data map into multiple smaller maps, locking only the shard in use. 
