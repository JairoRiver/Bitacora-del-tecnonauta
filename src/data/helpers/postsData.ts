import postsData from '../posts.json';

// 1. Definimos la interfaz para tener autocompletado y seguridad
interface Post {
    Title: string;
    Categories: string[];
    [key: string]: any; // Permite el resto de campos del JSON, por el momento no necesitamos más
}

/**
 * Obtiene una lista de categorías únicas a partir de un array de posts.
 */
export function getCategories(posts: Post[] = postsData): string[] {
    const allCategories = posts.flatMap((post) => post.Categories);

    const uniqueCategories = [...new Set(allCategories)];

    return uniqueCategories.sort((a, b) => a.localeCompare(b));
}
