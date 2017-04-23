* Add enough tests to not be embarrassed
* Implement missing features:
  + other lookup sources
* Full strings of characters, and transformations?
* switch generated unicode data to be structured instead of parsed at startup
* switch CharInfo to be an interface, so that stuff like regional indicators
  can supply another object, satisfying that interface but also others; lets
  the default Unicode characters not have widths, etc.

### Known Issues

Search via `character named -/` is a bit iffy on multiple words.  I need to
get more familiar with Ferret.
