* Add enough tests to not be embarrassed
* Implement missing features:
  + other lookup sources
* Full strings of characters, and transformations?
* switch generated unicode data to be structured instead of parsed at startup
* switch CharInfo to be an interface, so that stuff like regional indicators
  can supply another object, satisfying that interface but also others; lets
  the default Unicode characters not have widths, etc.
* RFC 1345 mnemonic digraphs as a column
* Rework to generate fields and be able to loop to see if a field is present
  at all, so that the `Of` column can usually disappear.

### Known Issues

Search via `character named -/` is a bit iffy on multiple words.  I need to
get more familiar with Ferret.
