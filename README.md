# gincgcode
I have a desktop CNC router but do not have software capable of making GCODE from STL with roughing passes. This is a problem when the STL is deeper than I can cut in a single pass.
This program is an attempt to create "poor mans roughing passes" It converts finish gcode into multiple passes that progresivly cut more away.
Each pass leaves a minimum thickness of material for the final finish pass.
I wrote it for V bits but I should also work with ball nose and flat bits.
