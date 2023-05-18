select
    (select
         count(*)
     from
         candidate_actors ca)  as furries,
    (select
         count(*)
     from
         candidate_posts cp)   as furry_posts,
    (select
         count(*)
     from
         candidate_likes cl)   as furry_likes,
    (select
         count(*)
     from
         candidate_follows cl) as furry_follows;