select
	(select count(*) from candidate_repositories cr) as furries,
	(select count(*) from candidate_posts cp ) as furry_posts,
	(select count(*) from candidate_likes cl) as furry_likes;