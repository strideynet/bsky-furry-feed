select
    (
        select count(*)
        from
            candidate_actors
    ) as furries,
    (
        select count(*)
        from
            candidate_posts
    ) as furry_posts,
    (
        select count(*)
        from
            candidate_likes
    ) as furry_likes,
    (
        select count(*)
        from
            candidate_follows
    ) as furry_follows;
