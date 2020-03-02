package main

import (
	"context"

	"cloud.google.com/go/firestore"
)

func findActors(ctx context.Context, c *firestore.Client) ([]Actor, error) {
	docs, err := c.Collection("Actor").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	actors := []Actor{}
	for _, doc := range docs {
		var a Actor
		doc.DataTo(&a)
		a.id = doc.Ref.ID
		actors = append(actors, a)
	}

	return actors, nil
}

func findActor(id string, actors []Actor) (Actor, error) {
	for _, a := range actors {
		if id == a.id {
			return a, nil
		}
	}

	return Actor{}, ErrNotFound
}

func (a Actor) update(ctx context.Context, c *firestore.Client) error {
	_, err := c.Collection("Actor").Doc(a.id).Set(ctx, a)
	return err
}
