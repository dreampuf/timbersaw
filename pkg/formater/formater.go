package formater

type Formater interface {
	Format(string) *Entity
	Put(*Entity)
}
