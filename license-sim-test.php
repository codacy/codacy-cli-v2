<?php

function wpmu_activate_stylesheet() {
	?>
	<style type="text/css">
		.wp-activate-container { width: 90%; margin: 0 auto; }
		.wp-activate-container form { margin-top: 2em; }
		#submit, #key { width: 100%; font-size: 24px; box-sizing: border-box; }
		#language { margin-top: 0.5em; }
		.wp-activate-container .error { background: #f66; color: #333; }
		span.h3 { padding: 0 8px; font-size: 1.3em; font-weight: 600; }
	</style>
	<?php
}
add_action( 'wp_head', 'wpmu_activate_stylesheet' );
add_action( 'wp_head', 'wp_strict_cross_origin_referrer' );
add_filter( 'wp_robots', 'wp_robots_sensitive_page' );

get_header( 'wp-activate' );

$blog_details = get_site();