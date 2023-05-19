select
    count(*) as follows,
    subject_did
from
    candidate_follows cf
where
        subject_did not in (
        select
            did
        from
            candidate_actors)
group by
    subject_did
order by
    follows desc