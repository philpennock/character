* Add enough tests to not be embarrassed
* Implement missing features:
  + other lookup sources
* Full strings of characters, and transformations?
* switch generated unicode data to be structured instead of parsed at startup
* switch CharInfo to be an interface, so that stuff like regional indicators
  can supply another object, satisfying that interface but also others; lets
  the default Unicode characters not have widths, etc.
* ~~RFC 1345 mnemonic digraphs as a column~~
  + have this now via X11 which is a superset of RFC 1345
* Rework to generate fields and be able to loop to see if a field is present
  at all, so that the `Of` column can usually disappear.
* Bits in a bitfield per CharInfo, representing attributes such as control
  + if marked as combining, combine with flag-specified char, by default SPACE.
    - will repair another column alignment bug
    - make sure to handle RTL combining, for 0x656 inter alia
* Emoji-mode is messing with display-width calculations, but is not always
  supported, so how do we determine those which are?
  + at what point do we surrender and have tabular use terminfo movements?

### Known Issues

#### Search

Search via `character named -/` is a bit iffy on multiple words.  I need to
get more familiar with Ferret.

#### CJK

While Hiragana and Katakana are supported, Kanji are not.
The supported characters are those identified with a name in the Unicode text
documents; the PDFs with sample glyphs are not scanned or isolated.
I'd like to get lists of known Hiragana values for each Kanji, and offer those
up, but this will be Japanese-specific whereas Unicode just treats these as
CJK (with the ensuing political mess that creates).
So not only do we need new table forms for showing these characters,
but we also need to have language-specific variants (or show each language, in
columns).

In the meantime, perhaps the covered defined ranges should be handled and
shown nameless, but with a glyph and codepoints, rather than claiming that the
codepoints are unknown.

At present, `character name 三` yields: `looking up "三": unknown codepoint 4e09`

At the very least, we should be able to treat that as not-an-error, since
we know that 0x4E09 is in a defined range (especially in this case, where the
character was provided as input).

Eg:

```
┏━━━┳━━━━━━━━━━━━┳━━━━━━┳━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━┳━━━━━━┳━━━━━┳━━━━┓
┃ C ┃ Name       ┃  Hex ┃   Dec ┃ Block                  ┃ Vim ┃ HTML ┃ XML ┃ Of ┃
┣━━━╇━━━━━━━━━━━━╇━━━━━━╇━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━╇━━━━━━╇━━━━━╇━━━━┫
┃ 三│            │ 4E09 │ 19977 │ CJK Unified Ideographs │     │      │     │    ┃
┗━━━┷━━━━━━━━━━━━┷━━━━━━┷━━━━━━━┷━━━━━━━━━━━━━━━━━━━━━━━━┷━━━━━┷━━━━━━┷━━━━━┷━━━━┛
```

Or ideally:

```
┏━━━┳━━━━━━┳━━━━━━━┳━━━━━━━┳━━━━━━━━━┳━━━━━━━━┓
┃ C ┃  Hex ┃   Dec ┃ Kanji ┃ K Count ┃ JIS    ┃ # other columns here for other languages
┣━━━╇━━━━━━╇━━━━━━━╇━━━━━━━╇━━━━━━━━━╇━━━━━━━━┫
┃ 三│ 4E09 │ 19977 │ さん  │     118 │ 0 3B30 ┃
┃   │      │       │ み    │      50 │        ┃
┃   │      │       │ ぞう  │      41 │        ┃
┃   │      │       │ さぶ  │       9 │        ┃
┃   │      │       │ みつ  │       8 │        ┃
┃   │      │       │ さ    │       1 │        ┃
┗━━━┷━━━━━━┷━━━━━━━┷━━━━━━━┷━━━━━━━━━┷━━━━━━━━┛
```

Could be multiple columns, one per language, or a `character japanese`
command, and one each for other languages?  I really don't have the expertise
to commit to that.

See <https://en.wiktionary.org/wiki/Category:Japanese_terms_spelled_with_%E4%B8%89>
for those numbers; Wikipedia has an API, could use it programmatically to
regenerate Go source.  The J0 number comes from Unicode's `CodeCharts.pdf`,
which has the following identifiers: `G0-487D HB1-A454 T1-4435 J0-3B30 K0-5F32 V1-4A26`.
We'd need at least 4 columns in this case, more if each C variant gets its own
column.
