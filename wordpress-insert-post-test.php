<?php
/**
 * Insert a new blog post into the database.
 *
 * @global wpdb $wpdb WordPress database abstraction object.
 *
 * @param string $post_title   The post title.
 * @param string $post_content The post content.
 * @param int    $post_author  The user ID of the author.
 * @param string $post_status  The post status (default: 'publish').
 * @return int|false The inserted post ID on success, false on failure.
 */
function wp_insert_blog_post( $post_title, $post_content, $post_author, $post_status = 'publish' ) {
	global $wpdb;

	// Prepare post data array.
	$post_data = array(
		'post_author'   => $post_author,
		'post_title'    => $post_title,
		'post_content'  => $post_content,
		'post_status'   => $post_status,
		'post_type'     => 'post',
		'post_date'     => current_time( 'mysql' ),
		'post_date_gmt' => current_time( 'mysql', 1 ),
	);

	// Insert the post into the database.
	$inserted = $wpdb->insert(
		$wpdb->posts,
		$post_data
	);

	if ( $inserted ) {
		return $wpdb->insert_id;
	}

	return false;
}

// Example usage (for testing):
// This will only work in a WordPress environment.
if ( function_exists( 'current_time' ) && isset( $GLOBALS['wpdb'] ) ) {
	$post_id = wp_insert_blog_post( 'Test Post', 'This is the content of the test post.', 1 );
	if ( $post_id ) {
		echo 'Inserted post with ID: ' . $post_id . "\n";
	} else {
		echo "Failed to insert post.\n";
	}
} else {
	echo "This script must be run within a WordPress environment.\n";
} 