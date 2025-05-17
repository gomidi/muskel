# Idea

general improvement and simplicifation of the problem that later items cut previous ones and may even "go back in time"

## concept

broadly spoken, when it comes to one item being stopped by the next, we have 5 categories of items:

1. RestItem/NoteItem
2. PatternItem/PartsItem/RepetionItem
3. MidiSampleItem
4. NoteOnItem/NoteOffItem
5. NotePatternOverwriteItem/RestPatternOverwriteItem

(1) Basically one item of a category is stopped by another one of the same category or another category.
Categories 4 and 5 are special, but apart from them, any item of the categories 1 to 3 can be stopped by another item of these
categories.

(2) On top of it, NoteOnItems can also stop any item of category 1 - 3, also themselves, but can only be stopped by NoteOffItems, and
items of category 2 and 3.

(3) Items of category 5 won't stop anything, but overwrite notes at the same position in the underlying pattern.

What this means, is that we can handle (1) and (2) in a general way.
This concept is a POC how it could be done.

For (3) it is suggested, that we unify patterns, parts and repetions in some way and then while parsing pass them any overwrites with their position. The here suggested `PositionedItem#Stream` method gets the `nextPatternStopItem` and the `endPos` of the piece.
These can be used to only calculate the pattern as long as needed. While the stopping mechanic is still "refining" the stopping and throwing away. 

So this basically means that items category 5 are not treated at all when producing the SMF, because they are already converted to
category 1 and category 4 items from the pattern and the pattern just returns them as part of the events stream.

For the stopping here is a special meta message defined (which might change). It is at the moment a smf.MetaSequencerData message of
value []byte{0x7f, 0x7f}. After a call to GetTrackEvents() the resulting stream events just contain midi messages, but still the information, if a noteon message should be tracked (NoteOnItem messages should not, everything else: yes). So every time
the stopping meta message occurs, it will be replaced by midi note off messages for all tracked midi note on messages that are still "on", followed by a pitchbend reset message (middle value). at the end of the piece there should be a general midi all notes off message to stop
NoteOnItem messages that were not stopped by the NoteOffItem counterpart.

Have a look at the tests and the calc file.