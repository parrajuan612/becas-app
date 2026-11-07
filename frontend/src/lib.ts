export const cop = (v: number) => '$ ' + v.toLocaleString('es-CO')

export const fechaCorta = (iso: string) => {
  const [y, m, d] = iso.split('-')
  return `${d}/${m}/${y}`
}
