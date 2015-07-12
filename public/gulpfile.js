var gulp       = require("gulp"),
    babel      = require("babelify"),
    sourcemaps = require("gulp-sourcemaps"),
    browserify = require("browserify"),
    source     = require("vinyl-source-stream"),
    buffer     = require("vinyl-buffer"),
    watchify   = require("watchify"),
    es         = require("event-stream"),
    glob       = require("glob"),
    path       = require("path");

var JS_SOURCE_DIR = "./src/js/";

function compileJS(watch, done) {
	glob(JS_SOURCE_DIR + "**/*.js", function (err, files) {
		if (err) done(err);

		var tasks = files.map(function (entry) {
			var b = browserify({
				entries: [entry]
			}).transform(babel);

			if (watch) {
				b = watchify(b);

				b.on("update", function () {
					console.log("File change detected. Recompiling...");

					return b.bundle()
						.on("error", function (err) { console.error(err); this.emit("end"); })
						.pipe(source(path.basename(entry)))
						.pipe(buffer())
						.pipe(sourcemaps.init({loadMaps: true}))
						.pipe(sourcemaps.write("."))
						.pipe(gulp.dest("dist/js"));
				})
			}

			// for now we'll have to do with duplicating this bundle call
			// because if I move this to a function, it somehow crashes :(
			return b.bundle()
				.on("error", function (err) { console.error(err); this.emit("end"); })
				.pipe(source(path.basename(entry)))
				.pipe(buffer())
				.pipe(sourcemaps.init({loadMaps: true}))
				.pipe(sourcemaps.write("."))
				.pipe(gulp.dest("dist/js"));
		});

		es.merge(tasks).on("end", done);
	});
}

gulp.task("javascript", function (done) {
	return compileJS(false, done);
})

gulp.task("watch", function (done) {
	return compileJS(true, done);
})