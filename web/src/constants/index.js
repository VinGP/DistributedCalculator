const BASE_API_URL = "https://api.schedule.vingp.dev/api/v1"

const COURSES_URL = `${BASE_API_URL}/schedule/courses?faculty=`
const FACULTIES_URL = `${BASE_API_URL}/schedule/faculties`
const GROUP_URL = `${BASE_API_URL}/schedule/groups/`

const DAY_URL = `${BASE_API_URL}/schedule/day`
const GROUPS_URL = `${BASE_API_URL}/schedule/groups`
const getGroupUrl = (group) => `${GROUP_URL}${group}`
const getGroupsUrl = (faculty, course) => `${GROUPS_URL}?faculty=${faculty}&course=${course}`

const getCourseUrl = (faculty) => `${COURSES_URL}${faculty}`

export {COURSES_URL, FACULTIES_URL, GROUP_URL, GROUPS_URL, DAY_URL, getGroupUrl, getGroupsUrl, getCourseUrl}