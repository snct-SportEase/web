import DOMPurify from 'isomorphic-dompurify';

const DEFAULT_OPTIONS = {
	ADD_ATTR: ['style', 'target', 'rel'],
	ADD_TAGS: ['iframe']
};

export function sanitizeHtml(html) {
	if (!html) {
		return '';
	}
	return DOMPurify.sanitize(html, DEFAULT_OPTIONS);
}


