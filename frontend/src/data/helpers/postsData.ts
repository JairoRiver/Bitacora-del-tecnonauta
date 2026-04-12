// Base URL of the Go backend — set via API_URL in .env at build time.
const API_URL = import.meta.env.API_URL ?? 'http://localhost:8080';

// ── Type exports (used by components) ────────────────────────────────────────

export type ParagraphType = 'p';
export type CodeType = 'c';
export type ImageType = 'i';
export type TableType = 't';
export type ParagraphStyleType = 'size-lg' | 'bold' | 'italic' | 'underline' | 'color-1' | 'color-2' | 'color-3';
export type CellStyleType = 'size-lg' | 'size-md' | 'size-xs' | 'bold' | 'italic' | 'underline';
export type LanguageType = 'SQL' | 'TypeScript' | 'Go' | 'Python' | 'Rust';
export type CodeStyleType = 'color-1' | 'color-2' | 'color-3';
export type CodeValueType = { language: LanguageType; text: string };
export type ImageValueType = { src: string; caption: string };
export type ColumnStyle = { header: string | number; style: CellStyleType[]; colored_numbers?: boolean };
export type TableConf = { has_header: boolean; table_styles: CellStyleType[]; column_style: ColumnStyle[] };
export type TableValueType = (string | number)[][];

interface TextBlock {
    type: ParagraphType;
    value: string;
    conf: ParagraphStyleType[];
}
interface CodeBlock {
    type: CodeType;
    value: CodeValueType;
    conf: CodeStyleType;
}
interface ImageBlock {
    type: ImageType;
    value: ImageValueType;
}
interface TableBlock {
    type: TableType;
    value: TableValueType;
    conf: TableConf;
}
type ContentBlock = TextBlock | CodeBlock | ImageBlock | TableBlock;

interface BlogPost {
    Title: string;
    Subtitle: string;
    Date: string;
    Categories: string[];
    HeroImage: string;
    Slug: string;
    Content: ContentBlock[];
}

interface CardData {
    title: string;
    description: string;
    date: string;
    categories: string[];
    heroImage: string;
    slug: string;
}

interface BlogPostList {
    Title: string;
    Slug: string;
}

// ── API response types (snake_case from Go backend) ───────────────────────────

interface ApiPostSummary {
    id: string;
    title: string;
    subtitle: string;
    date: string;
    hero_image: string;
    slug: string;
    categories: string[];
}

interface ApiPostDetail extends ApiPostSummary {
    content: ContentBlock[];
}

interface ApiCategory {
    name: string;
    slug: string;
}

// ── Raw fetch helpers ─────────────────────────────────────────────────────────

async function fetchPosts(): Promise<ApiPostSummary[]> {
    const res = await fetch(`${API_URL}/api/posts`);
    if (!res.ok) throw new Error(`GET /api/posts failed: ${res.status}`);
    return res.json();
}

async function fetchPost(slug: string): Promise<ApiPostDetail> {
    const res = await fetch(`${API_URL}/api/posts/${slug}`);
    if (!res.ok) throw new Error(`GET /api/posts/${slug} failed: ${res.status}`);
    return res.json();
}

async function fetchCategories(): Promise<ApiCategory[]> {
    const res = await fetch(`${API_URL}/api/categories`);
    if (!res.ok) throw new Error(`GET /api/categories failed: ${res.status}`);
    return res.json();
}

// ── Public API ────────────────────────────────────────────────────────────────

export function parseRouteName(name: string): string {
    return name.toLowerCase().replace(/\s+/g, '-');
}

export async function getCategories(): Promise<string[]> {
    const cats = await fetchCategories();
    return cats.map((c) => c.name).sort((a, b) => a.localeCompare(b));
}

export async function getCardInfo(): Promise<CardData[]> {
    const posts = await fetchPosts();
    return posts.map((p) => ({
        title: p.title,
        description: p.subtitle,
        date: p.date,
        categories: p.categories,
        heroImage: p.hero_image,
        slug: p.slug,
    }));
}

export async function getPostsList(): Promise<BlogPostList[]> {
    const posts = await fetchPosts();
    return posts.map((p) => ({ Title: p.title, Slug: p.slug }));
}

export async function getBlogPost(slug: string): Promise<BlogPost> {
    const p = await fetchPost(slug);
    return {
        Title: p.title,
        Subtitle: p.subtitle,
        Date: p.date,
        Categories: p.categories,
        HeroImage: p.hero_image,
        Slug: p.slug,
        Content: p.content,
    };
}
