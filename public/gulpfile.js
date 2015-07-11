var gulp  = require("gulp"),
    babel = require("gulp-babel");

gulp.task("default", function () {
	return gulp.src("src/js/post-editor.js")
	           .pipe(babel())
	           .pipe(gulp.dest("dist/js"));
})