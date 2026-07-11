-- Class representative roles are no longer used for authorization.
DELETE ur
FROM user_roles ur
INNER JOIN roles r ON r.id = ur.role_id
WHERE RIGHT(r.name, 4) = '_rep';

-- Remove the now-unreferenced role definitions as well.
DELETE r
FROM roles r
LEFT JOIN user_roles ur ON ur.role_id = r.id
WHERE RIGHT(r.name, 4) = '_rep'
  AND ur.role_id IS NULL;
