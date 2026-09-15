export function toDegrees(rad) {
    return rad * 180 / Math.PI;
}
export function toRadians(deg) {
    return deg * Math.PI / 180;
}
export function subtract(p1, p2) {
    return { x: p1.x - p2.x, y: p1.y - p2.y };
}
export function angleBetween(p) {
    return Math.atan2(p.y, p.x);
}
export function normalizeAngle(angle) {
    return (angle + 2 * Math.PI) % (2 * Math.PI);
}
export function computeRotationAngle(A, B, C, center) {
    const Arel = subtract(A, center);
    const Brel = subtract(B, center);
    const Crel = subtract(C, center);
    const AB = subtract(Brel, Arel);
    const AC = subtract(Crel, Arel);
    const angleAB = angleBetween(AB);
    const angleAC = angleBetween(AC);
    const rotation = normalizeAngle(angleAC - angleAB);
    return rotation;
}
export function rotateAround(point, center, angleRad) {
    const translated = subtract(point, center);
    const cos = Math.cos(angleRad);
    const sin = Math.sin(angleRad);
    const rotated = {
        x: translated.x * cos - translated.y * sin,
        y: translated.x * sin + translated.y * cos
    };
    return {
        x: rotated.x + center.x,
        y: rotated.y + center.y
    };
}
