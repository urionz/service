package cache

type TaggedCache struct {
	Repository
	tags *TagSet
}

type TaggableStore struct {
	BaseStore
}

type ITaggableStore interface {
	IRepository
}

func NewTaggedCache(store Store, tags *TagSet) *TaggedCache {
	taggedCache := &TaggedCache{
		tags: tags,
	}
	taggedCache.store = store
	return taggedCache
}

func (tag *TaggableStore) Tags(names ...string) (ITaggableStore, error) {
	return NewTaggedCache(tag, NewTagSet(tag, names...)), nil
}
