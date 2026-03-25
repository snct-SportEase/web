import DOMPurify from 'isomorphic-dompurify';

const DEFAULT_OPTIONS = {
	ADD_ATTR: ['target', 'rel'],
};

export function sanitizeHtml(html) {
	if (!html) {
		return '';
	}
	return DOMPurify.sanitize(html, DEFAULT_OPTIONS);
}


