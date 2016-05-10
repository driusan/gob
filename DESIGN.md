# The Gob Architecture

If you'd like to contribute, a brief overview of the architecture might help you get 
started.

Generally, the css package is responsible for parsing CSS and determining if a rule matches an
html.Node element. It can probably be used independently of Gob. it's responsible for
matching styles to elements, and dealing with what the spec calls "the specified value." 
It doesn't know anything about the rendering environment (such as the viewport size or containing 
block size, the percentage basis for many property types, etc) so it doesn't have enough information 
to calculate the "calculated value," but there are ConvertXtypeToY helper functions which take
a specified value and the needed information as parameters, and calculate the native Go type.

The renderer package in general, and the RenderableDomElement type in particular, does most of
the heavy lifting of calculating the calculated values, and then rendering them to an image.Image.
There's generally a GetX() receiver function for any corresponding CSS property that's implemented
which wraps the css package and Convert functions to get the calculated value.

The Render method generally works by recursively calling Draw on children of the current HTML element,
and moving an image.Point called "dot", which represents the top-left corner of the next child to draw. 
Dot is advanced appropriately after drawing, mostly based on the display property of the child (but 
sometimes affected by other properties...) It does two passes, one that doesn't draw to calculate the 
size of the RGBA image that needs to be allocated for drawing into, and one that does the drawing itself.

The net package is a thin wrapper around http.Client. It caches into ~/.gob/cache, and loads from the
cache if available (there's currently no cache invalidation, and it will always prefer the cache if
possible). It also fakes the User-Agent string so that web sites that do user-agent sniffing
give us real content.

The dom package isn't used yet, but will likely have more content once javascript support is added. It's
mostly just a placeholder for having a typedef for html.Node that we can add methods to.

Finally, the main package of Gob contains the shiny glue that puts it all together and provides the UI
around it.
into a RenderableDomElement, handles keyboard/mouse i/o, 
